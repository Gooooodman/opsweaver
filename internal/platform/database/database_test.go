package database_test

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Gooooodman/opsweaver/internal/platform/database"
)

const (
	envIntegration = "OPSWEAVER_INTEGRATION_TESTS"
	envServerDSN   = "OPSWEAVER_SERVER_DATABASE_DSN"
	envGatewayDSN  = "OPSWEAVER_GATEWAY_DATABASE_DSN"
)

func skipIfNoIntegration(t *testing.T) {
	t.Helper()
	if os.Getenv(envIntegration) != "1" {
		t.Skipf("set %s=1 to run integration tests", envIntegration)
	}
}

func requireDSN(t *testing.T, envKey string) string {
	t.Helper()
	dsn := os.Getenv(envKey)
	if dsn == "" {
		t.Skipf("%s not set, skipping", envKey)
	}
	return dsn
}

func TestOpen_ServerDB_Pings(t *testing.T) {
	skipIfNoIntegration(t)
	dsn := requireDSN(t, envServerDSN)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := database.Open(ctx, database.Options{DSN: dsn})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() {
		if err := database.Close(db); err != nil {
			t.Errorf("Close: %v", err)
		}
	})

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB: %v", err)
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		t.Fatalf("PingContext: %v", err)
	}
}

func TestOpen_GatewayDB_Pings(t *testing.T) {
	skipIfNoIntegration(t)
	dsn := requireDSN(t, envGatewayDSN)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := database.Open(ctx, database.Options{DSN: dsn})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() {
		if err := database.Close(db); err != nil {
			t.Errorf("Close: %v", err)
		}
	})

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB: %v", err)
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		t.Fatalf("PingContext: %v", err)
	}
}

func TestOpen_TwoDSNsAreIndependent(t *testing.T) {
	skipIfNoIntegration(t)
	serverDSN := requireDSN(t, envServerDSN)
	gatewayDSN := requireDSN(t, envGatewayDSN)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	serverDB, err := database.Open(ctx, database.Options{DSN: serverDSN})
	if err != nil {
		t.Fatalf("Open server: %v", err)
	}
	t.Cleanup(func() {
		if err := database.Close(serverDB); err != nil {
			t.Errorf("Close server: %v", err)
		}
	})

	gatewayDB, err := database.Open(ctx, database.Options{DSN: gatewayDSN})
	if err != nil {
		t.Fatalf("Open gateway: %v", err)
	}
	t.Cleanup(func() {
		if err := database.Close(gatewayDB); err != nil {
			t.Errorf("Close gateway: %v", err)
		}
	})

	// Random suffix so reruns and parallel test runs do not collide.
	suffix := fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Int63())
	tableName := fmt.Sprintf("_opsweaver_independence_test_%s", suffix)

	createSQL := fmt.Sprintf("CREATE TABLE %s (id INT PRIMARY KEY)", tableName)
	if err := serverDB.WithContext(ctx).Exec(createSQL).Error; err != nil {
		t.Fatalf("create table on server db: %v", err)
	}
	t.Cleanup(func() {
		dropSQL := fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)
		if err := serverDB.Exec(dropSQL).Error; err != nil {
			t.Errorf("drop table on server db: %v", err)
		}
	})

	insertSQL := fmt.Sprintf("INSERT INTO %s (id) VALUES (1)", tableName)
	if err := serverDB.WithContext(ctx).Exec(insertSQL).Error; err != nil {
		t.Fatalf("insert into server db: %v", err)
	}

	querySQL := fmt.Sprintf("SELECT id FROM %s", tableName)
	var id int
	err = gatewayDB.WithContext(ctx).Raw(querySQL).Scan(&id).Error
	if err == nil {
		t.Fatalf("expected gateway db query to fail because table %q exists only on server db, but it succeeded with id=%d", tableName, id)
	}
	if !strings.Contains(strings.ToLower(err.Error()), "does not exist") {
		t.Fatalf("expected 'does not exist' error from gateway db, got: %v", err)
	}
}

func TestOpen_InvalidDSN_ReturnsError(t *testing.T) {
	skipIfNoIntegration(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Port 1 is privileged and not listening; ping must fail.
	dsn := "postgres://opsweaver:opsweaver@127.0.0.1:1/nonexistent?sslmode=disable&connect_timeout=2"
	db, err := database.Open(ctx, database.Options{DSN: dsn})
	if err == nil {
		_ = database.Close(db)
		t.Fatal("expected Open to fail with unreachable DSN, got nil error")
	}
	if db != nil {
		t.Fatalf("expected nil *gorm.DB on failure, got %#v", db)
	}
}

func TestOpen_RespectsConnectionPoolDefaults(t *testing.T) {
	skipIfNoIntegration(t)
	dsn := requireDSN(t, envServerDSN)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t.Run("default max open conns", func(t *testing.T) {
		db, err := database.Open(ctx, database.Options{DSN: dsn})
		if err != nil {
			t.Fatalf("Open: %v", err)
		}
		t.Cleanup(func() { _ = database.Close(db) })

		sqlDB, err := db.DB()
		if err != nil {
			t.Fatalf("db.DB: %v", err)
		}
		stats := sqlDB.Stats()
		if stats.MaxOpenConnections != 10 {
			t.Errorf("MaxOpenConnections = %d, want 10", stats.MaxOpenConnections)
		}
	})

	t.Run("explicit max open conns", func(t *testing.T) {
		db, err := database.Open(ctx, database.Options{DSN: dsn, MaxOpenConns: 5})
		if err != nil {
			t.Fatalf("Open: %v", err)
		}
		t.Cleanup(func() { _ = database.Close(db) })

		sqlDB, err := db.DB()
		if err != nil {
			t.Fatalf("db.DB: %v", err)
		}
		stats := sqlDB.Stats()
		if stats.MaxOpenConnections != 5 {
			t.Errorf("MaxOpenConnections = %d, want 5", stats.MaxOpenConnections)
		}
	})
}

// TestClose_NilDB documents Close behavior on nil input. It is safe to run without
// integration credentials because it does not touch the network.
func TestClose_NilDB(t *testing.T) {
	if err := database.Close(nil); err == nil {
		t.Fatal("expected non-nil error from Close(nil), got nil")
	} else if !errors.Is(err, database.ErrNilDB) {
		t.Errorf("Close(nil) = %v, want ErrNilDB", err)
	}
}

