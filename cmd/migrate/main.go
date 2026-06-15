// Command migrate applies versioned SQL migrations to the OpsWeaver databases.
//
// Usage:
//
//	migrate -db server|gateway -command up|down|version [-steps N]
//
// Database DSNs are read from the environment:
//
//	OPSWEAVER_SERVER_DATABASE_DSN
//	OPSWEAVER_GATEWAY_DATABASE_DSN
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/Gooooodman/opsweaver/migrations"
)

const (
	dbServer  = "server"
	dbGateway = "gateway"

	cmdUp      = "up"
	cmdDown    = "down"
	cmdVersion = "version"

	envServerDSN  = "OPSWEAVER_SERVER_DATABASE_DSN"
	envGatewayDSN = "OPSWEAVER_GATEWAY_DATABASE_DSN"
)

type options struct {
	db      string
	command string
	steps   int
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("migrate: ")

	opts, err := parseFlags(os.Args[1:], os.Stderr)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}

	if err := run(opts); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

// parseFlags parses CLI arguments into options and validates them.
// It is exported via lowercase to enable unit tests in the same package.
func parseFlags(args []string, errOut io.Writer) (options, error) {
	fs := flag.NewFlagSet("migrate", flag.ContinueOnError)
	fs.SetOutput(errOut)
	fs.Usage = func() {
		fmt.Fprintln(errOut, "Usage:")
		fmt.Fprintln(errOut, "  migrate -db server|gateway -command up|down|version [-steps N]")
		fmt.Fprintln(errOut)
		fmt.Fprintln(errOut, "Flags:")
		fs.PrintDefaults()
	}

	var opts options
	fs.StringVar(&opts.db, "db", "", "target database: server|gateway (required)")
	fs.StringVar(&opts.command, "command", cmdUp, "migration command: up|down|version")
	fs.IntVar(&opts.steps, "steps", 0, "number of steps for up|down; 0 means apply all")

	if err := fs.Parse(args); err != nil {
		return options{}, err
	}

	switch opts.db {
	case dbServer, dbGateway:
	case "":
		return options{}, errors.New("missing required flag: -db")
	default:
		return options{}, fmt.Errorf("invalid -db value %q, expected server|gateway", opts.db)
	}

	switch opts.command {
	case cmdUp, cmdDown, cmdVersion:
	default:
		return options{}, fmt.Errorf("invalid -command value %q, expected up|down|version", opts.command)
	}

	if opts.steps < 0 {
		return options{}, fmt.Errorf("invalid -steps value %d, must be >= 0", opts.steps)
	}

	if opts.command == cmdVersion && opts.steps != 0 {
		return options{}, errors.New("-steps is not valid with -command version")
	}

	return opts, nil
}

func run(opts options) error {
	dsn, err := resolveDSN(opts.db)
	if err != nil {
		return err
	}

	source, err := sourceFor(opts.db)
	if err != nil {
		return err
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, dsn)
	if err != nil {
		return fmt.Errorf("initialize migrator for %s: %w", opts.db, err)
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			log.Printf("close source for %s: %v", opts.db, srcErr)
		}
		if dbErr != nil {
			log.Printf("close database for %s: %v", opts.db, dbErr)
		}
	}()

	return dispatch(m, opts)
}

// dispatch executes the requested migration command. It is separated from run
// so that the migration logic is unit-testable with a fake migrator.
func dispatch(m migrator, opts options) error {
	switch opts.command {
	case cmdUp:
		return runUp(m, opts.steps, opts.db)
	case cmdDown:
		return runDown(m, opts.steps, opts.db)
	case cmdVersion:
		return printVersion(m, opts.db)
	default:
		return fmt.Errorf("unhandled command %q", opts.command)
	}
}

func runUp(m migrator, steps int, db string) error {
	var err error
	if steps == 0 {
		err = m.Up()
	} else {
		err = m.Steps(steps)
	}
	return reportMigrationResult(err, db, "up")
}

func runDown(m migrator, steps int, db string) error {
	var err error
	if steps == 0 {
		err = m.Down()
	} else {
		err = m.Steps(-steps)
	}
	return reportMigrationResult(err, db, "down")
}

func reportMigrationResult(err error, db, label string) error {
	if errors.Is(err, migrate.ErrNoChange) {
		fmt.Printf("no change for %s\n", db)
		return nil
	}
	if err != nil {
		return fmt.Errorf("%s migrations on %s: %w", label, db, err)
	}
	fmt.Printf("%s migrations applied to %s\n", label, db)
	return nil
}

func printVersion(m migrator, db string) error {
	version, dirty, err := m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		fmt.Printf("%s version=0 dirty=false\n", db)
		return nil
	}
	if err != nil {
		return fmt.Errorf("read version for %s: %w", db, err)
	}
	fmt.Printf("%s version=%d dirty=%t\n", db, version, dirty)
	return nil
}

func resolveDSN(db string) (string, error) {
	var envKey string
	switch db {
	case dbServer:
		envKey = envServerDSN
	case dbGateway:
		envKey = envGatewayDSN
	default:
		return "", fmt.Errorf("unknown database %q", db)
	}
	dsn := os.Getenv(envKey)
	if dsn == "" {
		return "", fmt.Errorf("environment variable %s is required", envKey)
	}
	return dsn, nil
}

func sourceFor(db string) (source.Driver, error) {
	sub, err := selectSourceFS(db)
	if err != nil {
		return nil, err
	}
	src, err := iofs.New(sub, ".")
	if err != nil {
		return nil, fmt.Errorf("build migration source for %s: %w", db, err)
	}
	return src, nil
}

func selectSourceFS(db string) (fs.FS, error) {
	var dir string
	switch db {
	case dbServer:
		dir = "opsweaver_server"
	case dbGateway:
		dir = "opsweaver_gateway"
	default:
		return nil, fmt.Errorf("unknown database %q", db)
	}
	sub, err := fs.Sub(migrations.FS, dir)
	if err != nil {
		return nil, fmt.Errorf("select embedded migrations for %s: %w", db, err)
	}
	return sub, nil
}

// migrator captures the subset of *migrate.Migrate used by dispatch so the
// command dispatch can be tested without a real database.
type migrator interface {
	Up() error
	Down() error
	Steps(n int) error
	Version() (uint, bool, error)
}
