package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/Gooooodman/opsweaver/internal/platform/config"
)

func TestMainWiresQueueServer(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "main.go", nil, 0)
	if err != nil {
		t.Fatalf("parse main.go: %v", err)
	}

	found := false
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "main" {
			continue
		}
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			ident, ok := call.Fun.(*ast.Ident)
			if ok && ident.Name == "newQueueServer" {
				found = true
			}
			return true
		})
	}
	if !found {
		t.Fatal("main() does not call newQueueServer")
	}
}

func TestMainStartsQueueServerSynchronously(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "main.go", nil, 0)
	if err != nil {
		t.Fatalf("parse main.go: %v", err)
	}

	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "main" {
			continue
		}
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			goStmt, ok := node.(*ast.GoStmt)
			if !ok {
				return true
			}
			ast.Inspect(goStmt, func(child ast.Node) bool {
				selector, ok := child.(*ast.SelectorExpr)
				if ok && selector.Sel.Name == "Start" {
					t.Fatal("main() starts the queue server in a goroutine; Start returns immediately")
				}
				return true
			})
			return true
		})
	}
}

func TestNewQueueServer(t *testing.T) {
	server, err := newQueueServer(config.RedisConfig{Addr: "localhost:6379", DB: 0})
	if err != nil {
		t.Fatalf("newQueueServer() error = %v", err)
	}
	if server == nil {
		t.Fatal("newQueueServer() = nil, want server")
	}
}
