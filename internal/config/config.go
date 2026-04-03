package config

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Config mirrors bom.yml structure needed by bomgen.
type Config struct {
	Output          OutputConfig
	IncludeTables   []string
	ExcludeTables   []string
	Dialect         string
	Alias           AliasConfig
	Associations    map[string][]Association
	AllowNullUnique bool
	Identity        map[string]map[string]string
}

type OutputConfig struct {
	Root string
}

type AliasConfig struct {
	Strategy string
	Width    int
}

type Association struct {
	Type        string
	To          string
	LocalKeys   []string
	ForeignKeys []string
}

// Parse reads bom.yml and returns the configured structure.
func Parse(r io.Reader) (Config, error) {
	var cfg Config
	cfg.Associations = make(map[string][]Association)
	cfg.Identity = make(map[string]map[string]string)

	scanner := bufio.NewScanner(r)
	block := ""
	var assocTable string
	var currentAssoc *Association
	var identityTable string

	for scanner.Scan() {
		line := scanner.Text()

		if trimmed := strings.TrimSpace(line); trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		indent := len(line) - len(strings.TrimLeft(line, " "))
		trimmed := strings.TrimSpace(line)

		switch {
		case indent == 0 && trimmed == "output:":
			block = "output"
			continue
		case indent == 0 && trimmed == "alias:":
			block = "alias"
			continue
		case indent == 0 && trimmed == "associations:":
			block = "associations"
			continue
		case indent == 0 && trimmed == "identity:":
			block = "identity"
			continue
		case indent == 0:
			block = ""
			assocTable = ""
			currentAssoc = nil
			identityTable = ""
		}

		switch block {
		case "output":
			if strings.HasPrefix(trimmed, "root:") {
				cfg.Output.Root = strings.TrimSpace(strings.TrimPrefix(trimmed, "root:"))
				continue
			}
			return cfg, fmt.Errorf("unknown output key: %q", trimmed)
		case "alias":
			if strings.HasPrefix(trimmed, "strategy:") {
				cfg.Alias.Strategy = strings.TrimSpace(strings.TrimPrefix(trimmed, "strategy:"))
				continue
			}
			if strings.HasPrefix(trimmed, "width:") {
				v := strings.TrimSpace(strings.TrimPrefix(trimmed, "width:"))
				n, err := strconv.Atoi(v)
				if err != nil {
					return cfg, fmt.Errorf("alias.width must be integer: %w", err)
				}
				cfg.Alias.Width = n
				continue
			}
			return cfg, fmt.Errorf("unknown alias key: %q", trimmed)
		case "associations":
			if indent == 2 {
				key, val := splitKeyValue(trimmed)
				assocTable = key
				if _, ok := cfg.Associations[assocTable]; !ok {
					cfg.Associations[assocTable] = nil
				}
				currentAssoc = nil
				if val == "[]" {
					cfg.Associations[assocTable] = nil
				}
				continue
			}
			if indent == 4 && strings.HasPrefix(trimmed, "-") {
				entry := Association{}
				cfg.Associations[assocTable] = append(cfg.Associations[assocTable], entry)
				currentAssoc = &cfg.Associations[assocTable][len(cfg.Associations[assocTable])-1]
				rest := strings.TrimSpace(strings.TrimPrefix(trimmed, "-"))
				if rest != "" {
					if err := parseAssocField(currentAssoc, rest); err != nil {
						return cfg, err
					}
				}
				continue
			}
			if indent >= 6 && currentAssoc != nil {
				if err := parseAssocField(currentAssoc, trimmed); err != nil {
					return cfg, err
				}
			}
		case "identity":
			if indent == 2 {
				key, val := splitKeyValue(trimmed)
				identityTable = key
				if val == "{}" || val == "" {
					cfg.Identity[identityTable] = make(map[string]string)
				}
				continue
			}
			if indent >= 4 && identityTable != "" {
				key, val := splitKeyValue(trimmed)
				if key != "" && val != "" {
					if cfg.Identity[identityTable] == nil {
						cfg.Identity[identityTable] = make(map[string]string)
					}
					cfg.Identity[identityTable][key] = val
				}
			}
		default:
			key, val := splitKeyValue(trimmed)
			switch key {
			case "include_tables":
				cfg.IncludeTables = parseInlineStrings(val)
			case "exclude_tables":
				cfg.ExcludeTables = parseInlineStrings(val)
			case "dialect":
				cfg.Dialect = val
			case "allow_null_unique":
				cfg.AllowNullUnique = parseBool(val)
			default:
				return cfg, fmt.Errorf("unknown top-level key: %q", key)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func parseAssocField(entry *Association, trimmed string) error {
	key, val := splitKeyValue(trimmed)
	switch key {
	case "type":
		entry.Type = val
	case "to":
		entry.To = val
	case "local_keys":
		entry.LocalKeys = parseInlineStrings(val)
	case "foreign_keys":
		entry.ForeignKeys = parseInlineStrings(val)
	default:
		return fmt.Errorf("unknown association key: %q", key)
	}
	return nil
}

func splitKeyValue(line string) (string, string) {
	parts := strings.SplitN(line, ":", 2)
	key := strings.TrimSpace(parts[0])
	val := ""
	if len(parts) == 2 {
		val = strings.TrimSpace(parts[1])
	}
	return key, val
}

func parseInlineStrings(value string) []string {
	if value == "" {
		return nil
	}
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
		raw := strings.TrimSpace(value[1 : len(value)-1])
		if raw == "" {
			return nil
		}
		parts := strings.Split(raw, ",")
		var out []string
		for _, part := range parts {
			if v := strings.TrimSpace(part); v != "" {
				out = append(out, trimQuotes(v))
			}
		}
		return out
	}
	return []string{trimQuotes(value)}
}

func parseBool(value string) bool {
	lower := strings.TrimSpace(strings.ToLower(value))
	return lower == "true" || lower == "1"
}

func trimQuotes(value string) string {
	value = strings.Trim(value, `"`)
	value = strings.Trim(value, `'`)
	return value
}
