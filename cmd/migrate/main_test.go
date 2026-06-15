package main

import (
	"bytes"
	"errors"
	"flag"
	"io/fs"
	"testing"

	"github.com/golang-migrate/migrate/v4"
)

func TestParseFlagsValid(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want options
	}{
		{
			name: "server up default",
			args: []string{"-db", "server"},
			want: options{db: "server", command: "up", steps: 0},
		},
		{
			name: "gateway down with steps",
			args: []string{"-db", "gateway", "-command", "down", "-steps", "2"},
			want: options{db: "gateway", command: "down", steps: 2},
		},
		{
			name: "server version",
			args: []string{"-db", "server", "-command", "version"},
			want: options{db: "server", command: "version", steps: 0},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			got, err := parseFlags(tc.args, &buf)
			if err != nil {
				t.Fatalf("parseFlags returned error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("parseFlags = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestParseFlagsInvalid(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{name: "missing db", args: []string{}},
		{name: "invalid db", args: []string{"-db", "bogus"}},
		{name: "invalid command", args: []string{"-db", "server", "-command", "sideways"}},
		{name: "negative steps", args: []string{"-db", "server", "-steps", "-1"}},
		{name: "steps with version", args: []string{"-db", "server", "-command", "version", "-steps", "1"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			if _, err := parseFlags(tc.args, &buf); err == nil {
				t.Fatalf("expected error, got nil; stderr=%q", buf.String())
			}
		})
	}
}

func TestParseFlagsHelpReturnsErrHelp(t *testing.T) {
	var buf bytes.Buffer
	_, err := parseFlags([]string{"-h"}, &buf)
	if !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("expected flag.ErrHelp, got %v", err)
	}
}

type fakeMigrator struct {
	upErr      error
	downErr    error
	stepsErr   error
	stepsCalls []int
	upCalls    int
	downCalls  int

	version      uint
	dirty        bool
	versionErr   error
	versionCalls int
}

func (f *fakeMigrator) Up() error {
	f.upCalls++
	return f.upErr
}

func (f *fakeMigrator) Down() error {
	f.downCalls++
	return f.downErr
}

func (f *fakeMigrator) Steps(n int) error {
	f.stepsCalls = append(f.stepsCalls, n)
	return f.stepsErr
}

func (f *fakeMigrator) Version() (uint, bool, error) {
	f.versionCalls++
	return f.version, f.dirty, f.versionErr
}

func TestDispatchUpAll(t *testing.T) {
	m := &fakeMigrator{}
	if err := dispatch(m, options{db: "server", command: "up", steps: 0}); err != nil {
		t.Fatalf("dispatch returned error: %v", err)
	}
	if m.upCalls != 1 {
		t.Fatalf("expected Up called once, got %d", m.upCalls)
	}
	if len(m.stepsCalls) != 0 {
		t.Fatalf("expected no Steps calls, got %v", m.stepsCalls)
	}
}

func TestDispatchUpSteps(t *testing.T) {
	m := &fakeMigrator{}
	if err := dispatch(m, options{db: "server", command: "up", steps: 3}); err != nil {
		t.Fatalf("dispatch returned error: %v", err)
	}
	if len(m.stepsCalls) != 1 || m.stepsCalls[0] != 3 {
		t.Fatalf("expected Steps(3), got %v", m.stepsCalls)
	}
	if m.upCalls != 0 {
		t.Fatalf("expected no Up calls, got %d", m.upCalls)
	}
}

func TestDispatchDownAll(t *testing.T) {
	m := &fakeMigrator{}
	if err := dispatch(m, options{db: "server", command: "down", steps: 0}); err != nil {
		t.Fatalf("dispatch returned error: %v", err)
	}
	if m.downCalls != 1 {
		t.Fatalf("expected Down called once, got %d", m.downCalls)
	}
}

func TestDispatchDownSteps(t *testing.T) {
	m := &fakeMigrator{}
	if err := dispatch(m, options{db: "server", command: "down", steps: 2}); err != nil {
		t.Fatalf("dispatch returned error: %v", err)
	}
	if len(m.stepsCalls) != 1 || m.stepsCalls[0] != -2 {
		t.Fatalf("expected Steps(-2), got %v", m.stepsCalls)
	}
}

func TestDispatchUpNoChangeTreatedAsSuccess(t *testing.T) {
	m := &fakeMigrator{upErr: migrate.ErrNoChange}
	if err := dispatch(m, options{db: "server", command: "up"}); err != nil {
		t.Fatalf("expected nil error for ErrNoChange, got %v", err)
	}
}

func TestDispatchUpErrorPropagated(t *testing.T) {
	want := errors.New("boom")
	m := &fakeMigrator{upErr: want}
	err := dispatch(m, options{db: "server", command: "up"})
	if err == nil || !errors.Is(err, want) {
		t.Fatalf("expected wrapped boom error, got %v", err)
	}
}

func TestDispatchVersionNilTreatedAsZero(t *testing.T) {
	m := &fakeMigrator{versionErr: migrate.ErrNilVersion}
	if err := dispatch(m, options{db: "gateway", command: "version"}); err != nil {
		t.Fatalf("expected nil error for ErrNilVersion, got %v", err)
	}
	if m.versionCalls != 1 {
		t.Fatalf("expected Version called once, got %d", m.versionCalls)
	}
}

func TestDispatchVersionSuccess(t *testing.T) {
	m := &fakeMigrator{version: 4, dirty: true}
	if err := dispatch(m, options{db: "server", command: "version"}); err != nil {
		t.Fatalf("dispatch returned error: %v", err)
	}
}

func TestResolveDSNMissingEnv(t *testing.T) {
	t.Setenv("OPSWEAVER_SERVER_DATABASE_DSN", "")
	if _, err := resolveDSN("server"); err == nil {
		t.Fatal("expected error when env not set")
	}
}

func TestResolveDSNFromEnv(t *testing.T) {
	t.Setenv("OPSWEAVER_GATEWAY_DATABASE_DSN", "postgres://example")
	dsn, err := resolveDSN("gateway")
	if err != nil {
		t.Fatalf("resolveDSN returned error: %v", err)
	}
	if dsn != "postgres://example" {
		t.Fatalf("dsn = %q, want postgres://example", dsn)
	}
}

func TestSelectSourceFSEmbedsEachDB(t *testing.T) {
	for _, db := range []string{"server", "gateway"} {
		t.Run(db, func(t *testing.T) {
			sub, err := selectSourceFS(db)
			if err != nil {
				t.Fatalf("selectSourceFS returned error: %v", err)
			}
			entries, err := fs.ReadDir(sub, ".")
			if err != nil {
				t.Fatalf("read embedded dir: %v", err)
			}
			if len(entries) < 2 {
				t.Fatalf("expected at least up+down sql for %s, got %d entries", db, len(entries))
			}
		})
	}
}
