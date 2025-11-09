package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"bom/internal/assoc"
	"bom/internal/codegen"
	"bom/internal/config"
	"bom/internal/parser/mysql"
	"bom/internal/schema"
)

func main() {
	var ddlPath = flag.String("ddl", "./schema.sql", "path to DDL file")
	var configPath = flag.String("config", "./bom.yml", "path to configuration")
	var outDir = flag.String("out", "./pkg/generated", "output directory for generated code")
	flag.Parse()

	ddl, err := os.ReadFile(*ddlPath)
	if err != nil {
		log.Fatalf("cannot read DDL: %v", err)
	}
	cfgData, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("cannot read config: %v", err)
	}

	cfg, err := config.Parse(bytes.NewReader(cfgData))
	if err != nil {
		log.Fatalf("config parse failed: %v", err)
	}

	parser := mysql.New()
	ir, err := parser.Parse(context.Background(), string(ddl))
	if err != nil {
		log.Fatalf("parse failed: %v", err)
	}
	filtered := filterTables(ir, cfg.IncludeTables, cfg.ExcludeTables)
	resolved, err := assoc.Resolve(filtered, cfg)
	if err != nil {
		log.Fatalf("assoc resolve: %v", err)
	}

	target := *outDir
	if cfg.Output.Root != "" {
		target = cfg.Output.Root
	}

	gen := codegen.New()
	if err := gen.Generate(resolved, target, cfg.Dialect); err != nil {
		log.Fatalf("codegen failed: %v", err)
	}

	fmt.Printf("bomgen finished: %d tables -> %s\n", len(resolved.Tables), target)
}

func filterTables(ir schema.IR, include, exclude []string) schema.IR {
	setInclude := make(map[string]struct{})
	for _, v := range include {
		setInclude[v] = struct{}{}
	}
	setExclude := make(map[string]struct{})
	for _, v := range exclude {
		setExclude[v] = struct{}{}
	}
	out := schema.New()
	for _, t := range ir.Tables {
		if len(setInclude) > 0 {
			if _, ok := setInclude[t.Name]; !ok {
				continue
			}
		}
		if _, ok := setExclude[t.Name]; ok {
			continue
		}
		out.AddTable(t)
	}
	return out
}
