package codegen

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"

	"bom/internal/schema"
)

//go:embed templates/findmany.tmpl
var findManyTemplate string

//go:embed templates/findunique.tmpl
var findUniqueTemplate string

// Generator produces Go code for the resolved schema.
type Generator struct {
	tmpl *template.Template
}

// New returns a generator configured with built-in templates.
func New() *Generator {
	funcs := template.FuncMap{
		"camel":      toCamel,
		"lowerFirst": lowerFirst,
	}
	combined := findManyTemplate + "\n" + findUniqueTemplate
	return &Generator{
		tmpl: template.Must(template.New("bom").Funcs(funcs).Parse(combined)),
	}
}

// Generate renders the templated {model,inputs,aliases} artifacts into outDir.
func (g *Generator) Generate(ir schema.IR, outDir string, dialectName string) error {
	if outDir == "" {
		return fmt.Errorf("outDir required")
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	dialectInfo, err := resolveDialect(dialectName)
	if err != nil {
		return err
	}

	data := struct {
		Tables             []tableData
		DialectImportPath  string
		DialectImportAlias string
		DialectInit        string
	}{
		Tables:             buildTables(ir),
		DialectImportPath:  dialectInfo.ImportPath,
		DialectImportAlias: dialectInfo.ImportAlias,
		DialectInit:        dialectInfo.Constructor,
	}

	var buf bytes.Buffer
	if err := g.tmpl.Execute(&buf, data); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(outDir, "generated.go"), buf.Bytes(), 0o644)
}

type tableData struct {
	TableName         string
	ModelName         string
	Columns           []columnData
	PrimaryKey        []columnData
	UniqueConstraints []uniqueConstraint
	UniqueInterface   string
	Relations         []relationData
}

type columnData struct {
	Name     string
	Camel    string
	GoType   string
	Nullable bool
}

type uniqueConstraint struct {
	TypeName string
	Fields   []uniqueField
}

type uniqueField struct {
	ColumnName string
	Camel      string
	GoType     string
}

type relationData struct {
	FieldName      string
	SelectArgsName string
	TargetModel    string
	TargetTable    string
	Kind           string
	LocalKeys      []string
	ForeignKeys    []string
	AllowSelectAll bool
}

type dialectInfo struct {
	ImportPath  string
	ImportAlias string
	Constructor string
}

func resolveDialect(name string) (dialectInfo, error) {
	dialects := map[string]dialectInfo{
		"mysql": {
			ImportPath:  "bom/pkg/dialect/mysql",
			ImportAlias: "dialectmysql",
			Constructor: "dialectmysql.New()",
		},
		"postgres": {
			ImportPath:  "bom/pkg/dialect/postgres",
			ImportAlias: "dialectpostgres",
			Constructor: "dialectpostgres.New()",
		},
		"sqlite": {
			ImportPath:  "bom/pkg/dialect/sqlite",
			ImportAlias: "dialectsqlite",
			Constructor: "dialectsqlite.New()",
		},
	}
	key := strings.ToLower(strings.TrimSpace(name))
	if key == "" {
		key = "mysql"
	}
	if info, ok := dialects[key]; ok {
		return info, nil
	}
	return dialectInfo{}, fmt.Errorf("unsupported dialect %q", name)
}

func buildTables(ir schema.IR) []tableData {
	var result []tableData
	modelNames := make(map[string]string)
	tableRefs := make(map[string]*schema.Table)
	for i := range ir.Tables {
		tbl := &ir.Tables[i]
		key := strings.ToLower(tbl.Name)
		modelNames[key] = toCamel(tbl.Name)
		tableRefs[key] = tbl
	}
	reachability := computeReachability(ir)
	for _, tbl := range ir.Tables {
		uniqueConstraints := buildUniqueConstraints(tbl)
		var uniqueInterface string
		if len(uniqueConstraints) > 0 {
			var parts []string
			for _, uc := range uniqueConstraints {
				parts = append(parts, uc.TypeName)
			}
			uniqueInterface = strings.Join(parts, " | ")
		}
		relations := buildRelations(tbl, modelNames, tableRefs, reachability)
		td := tableData{
			TableName:         tbl.Name,
			ModelName:         toCamel(tbl.Name),
			UniqueConstraints: uniqueConstraints,
			UniqueInterface:   uniqueInterface,
			Relations:         relations,
		}
		for _, col := range tbl.Columns {
			td.Columns = append(td.Columns, columnData{
				Name:     col.Name,
				Camel:    toCamel(col.Name),
				GoType:   col.GoType,
				Nullable: col.Nullable,
			})
		}
		for _, pk := range tbl.PrimaryKey {
			if col := findColumn(tbl.Columns, pk); col != nil {
				td.PrimaryKey = append(td.PrimaryKey, columnData{
					Name:   col.Name,
					Camel:  toCamel(col.Name),
					GoType: col.GoType,
				})
			}
		}
		result = append(result, td)
	}
	return result
}

func findColumn(cols []schema.Column, name string) *schema.Column {
	lower := strings.ToLower(name)
	for i := range cols {
		if strings.ToLower(cols[i].Name) == lower {
			return &cols[i]
		}
	}
	return nil
}

func toCamel(in string) string {
	var out strings.Builder
	upper := true
	for _, r := range in {
		switch {
		case r == '_' || r == '-':
			upper = true
		default:
			if upper {
				out.WriteRune(unicode.ToUpper(r))
			} else {
				out.WriteRune(unicode.ToLower(r))
			}
			upper = false
		}
	}
	if out.Len() == 0 {
		return strings.Title(in)
	}
	return out.String()
}

func lowerFirst(in string) string {
	if in == "" {
		return ""
	}
	runes := []rune(in)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

func buildUniqueConstraints(tbl schema.Table) []uniqueConstraint {
	var uniques []uniqueConstraint
	seen := make(map[string]struct{})
	addConstraint := func(cols []string) {
		if len(cols) == 0 {
			return
		}
		key := strings.ToLower(strings.Join(cols, "|"))
		if _, ok := seen[key]; ok {
			return
		}
		if uc, ok := makeUniqueConstraint(tbl, cols); ok {
			seen[key] = struct{}{}
			uniques = append(uniques, uc)
		}
	}
	if len(tbl.PrimaryKey) > 0 {
		addConstraint(tbl.PrimaryKey)
	}
	for _, uniq := range tbl.Uniques {
		addConstraint(uniq.Cols)
	}
	return uniques
}

func makeUniqueConstraint(tbl schema.Table, cols []string) (uniqueConstraint, bool) {
	var fields []uniqueField
	var typeParts []string
	for _, colName := range cols {
		col := findColumn(tbl.Columns, colName)
		if col == nil {
			return uniqueConstraint{}, false
		}
		camel := toCamel(col.Name)
		fields = append(fields, uniqueField{
			ColumnName: col.Name,
			Camel:      camel,
			GoType:     col.GoType,
		})
		typeParts = append(typeParts, camel)
	}
	if len(fields) == 0 {
		return uniqueConstraint{}, false
	}
	typeName := toCamel(tbl.Name) + "UK_" + strings.Join(typeParts, "")
	return uniqueConstraint{
		TypeName: typeName,
		Fields:   fields,
	}, true
}

func buildRelations(tbl schema.Table, modelNames map[string]string, tableRefs map[string]*schema.Table, reach map[string]map[string]bool) []relationData {
	var relations []relationData
	seen := make(map[string]struct{})
	tableModel := toCamel(tbl.Name)
	sourceKey := strings.ToLower(tbl.Name)
	for _, rel := range tbl.Relations {
		fieldName := rel.Name
		if fieldName == "" {
			fieldName = rel.To
		}
		fieldName = toCamel(fieldName)
		if fieldName == "" {
			continue
		}
		key := strings.ToLower(fieldName)
		if _, ok := seen[key]; ok {
			continue
		}
		targetKey := strings.ToLower(rel.To)
		targetModel := modelNames[targetKey]
		if targetModel == "" {
			targetModel = toCamel(rel.To)
		}
		if targetModel == "" {
			continue
		}
		targetTable := rel.To
		if t := tableRefs[targetKey]; t != nil {
			targetTable = t.Name
		}
		selectArgsName := fmt.Sprintf("%s%sSelectArgs", tableModel, fieldName)
		allowSelectAll := true
		if targetKey != "" {
			if reachable, ok := reach[targetKey]; ok {
				if reachable[sourceKey] {
					allowSelectAll = false
				}
			}
		}
		relations = append(relations, relationData{
			FieldName:      fieldName,
			SelectArgsName: selectArgsName,
			TargetModel:    targetModel,
			TargetTable:    targetTable,
			Kind:           rel.Kind,
			LocalKeys:      append([]string(nil), rel.LocalKeys...),
			ForeignKeys:    append([]string(nil), rel.ForeignKeys...),
			AllowSelectAll: allowSelectAll,
		})
		seen[key] = struct{}{}
	}
	return relations
}

func computeReachability(ir schema.IR) map[string]map[string]bool {
	graph := make(map[string][]string)
	nodes := make(map[string]struct{})
	for _, tbl := range ir.Tables {
		src := strings.ToLower(tbl.Name)
		nodes[src] = struct{}{}
		for _, rel := range tbl.Relations {
			dst := strings.ToLower(rel.To)
			graph[src] = append(graph[src], dst)
			nodes[dst] = struct{}{}
		}
	}
	reach := make(map[string]map[string]bool)
	for node := range nodes {
		visited := make(map[string]bool)
		dfsReach(node, graph, visited)
		reach[node] = visited
	}
	return reach
}

func dfsReach(node string, graph map[string][]string, visited map[string]bool) {
	for _, next := range graph[node] {
		if visited[next] {
			continue
		}
		visited[next] = true
		dfsReach(next, graph, visited)
	}
}
