package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"bom/internal/assoc"
	"bom/internal/codegen"
	"bom/internal/config"
	parseriface "bom/internal/parser"
	parsermysql "bom/internal/parser/mysql"
	parserpostgres "bom/internal/parser/postgres"
	parsersqlite "bom/internal/parser/sqlite"
	"bom/internal/schema"
)

func main() {
	var ddlPath = flag.String("ddl", "./schema.sql", "path to DDL file")
	var configPath = flag.String("config", "./bom.yml", "path to configuration")
	var outDir = flag.String("out", "./pkg/generated", "output directory for generated code")
	var parserOverride = flag.String("parser", "", "optional parser dialect override (mysql/postgres/sqlite)")
	flag.Parse()

	ddl, err := os.ReadFile(*ddlPath)
	if err != nil {
		log.Fatalf("cannot read DDL: %v", err)
	}
	const defaultConfigPath = "./bom.yml"
	var cfg config.Config
	cfgData, err := os.ReadFile(*configPath)
	if err != nil {
		if !(os.IsNotExist(err) && *configPath == defaultConfigPath) {
			log.Fatalf("cannot read config: %v", err)
		}
	} else {
		cfg, err = config.Parse(bytes.NewReader(cfgData))
		if err != nil {
			log.Fatalf("config parse failed: %v", err)
		}
	}

	parserName := cfg.Dialect
	if override := strings.TrimSpace(*parserOverride); override != "" {
		parserName = override
	}
	parser := selectParser(parserName)
	ir, err := parser.Parse(context.Background(), string(ddl))
	if err != nil {
		log.Fatalf("parse failed: %v", err)
	}
	filtered := filterTables(ir, cfg.IncludeTables, cfg.ExcludeTables)
	resolved, err := assoc.Resolve(filtered, cfg)
	if err != nil {
		log.Fatalf("assoc resolve: %v", err)
	}

	if err := applyIdentityConfig(&resolved, cfg.Identity); err != nil {
		log.Fatalf("identity config: %v", err)
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

func selectParser(name string) parseriface.DDLParser {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "sqlite":
		return parsersqlite.New()
	case "postgres":
		return parserpostgres.New()
	default:
		return parsermysql.New()
	}
}

func applyIdentityConfig(ir *schema.IR, identities map[string]map[string]string) error {
	if ir == nil || len(identities) == 0 {
		return nil
	}
	for tableName, cols := range identities {
		if cols == nil {
			continue
		}
		table := lookupTable(ir, tableName)
		if table == nil {
			return fmt.Errorf("identity: unknown table %q", tableName)
		}
		for colName, strategy := range cols {
			if idx := columnIndex(table.Columns, colName); idx >= 0 {
				table.Columns[idx].Identity = strings.TrimSpace(strings.ToLower(strategy))
			} else {
				return fmt.Errorf("identity: unknown column %q.%q", tableName, colName)
			}
		}
	}
	return nil
}

func lookupTable(ir *schema.IR, name string) *schema.Table {
	if ir == nil {
		return nil
	}
	lower := strings.ToLower(name)
	for i := range ir.Tables {
		if strings.ToLower(ir.Tables[i].Name) == lower {
			return &ir.Tables[i]
		}
	}
	return nil
}

func columnIndex(cols []schema.Column, name string) int {
	lower := strings.ToLower(name)
	for i := range cols {
		if strings.ToLower(cols[i].Name) == lower {
			return i
		}
	}
	return -1
}
