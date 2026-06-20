package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/Gooooodman/opsweaver/internal/platform/config"
)

func TestMainWiresQueueClient(t *testing.T) {
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
			if ok && ident.Name == "newQueueClient" {
				found = true
			}
			return true
		})
	}
	if !found {
		t.Fatal("main() does not call newQueueClient")
	}
}

func TestNewQueueClient(t *testing.T) {
	client, err := newQueueClient(config.RedisConfig{Addr: "localhost:6379", DB: 0})
	if err != nil {
		t.Fatalf("newQueueClient() error = %v", err)
	}
	if client == nil {
		t.Fatal("newQueueClient() = nil, want client")
	}
	if err := client.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}
