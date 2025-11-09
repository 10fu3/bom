package config

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	const raw = `
output:
  root: ./pkg/generated
include_tables: [video, actor]
exclude_tables: []
dialect: mysql
alias:
  strategy: base62
  width: 8
associations:
  video:
    - type: HasMany
      to: video_actor
      local_keys: [id]
      foreign_keys: [video_id]
  actor: []
`

	cfg, err := Parse(strings.NewReader(raw))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if cfg.Output.Root != "./pkg/generated" {
		t.Errorf("unexpected output root: %s", cfg.Output.Root)
	}
	if cfg.Dialect != "mysql" {
		t.Errorf("unexpected dialect %q", cfg.Dialect)
	}
	if cfg.Alias.Strategy != "base62" {
		t.Errorf("unexpected alias strategy %q", cfg.Alias.Strategy)
	}
	if cfg.Alias.Width != 8 {
		t.Errorf("unexpected width %d", cfg.Alias.Width)
	}
	if len(cfg.IncludeTables) != 2 {
		t.Errorf("expected include tables, got %v", cfg.IncludeTables)
	}
	if _, ok := cfg.Associations["video"]; !ok {
		t.Fatalf("video association missing")
	}
	if len(cfg.Associations["actor"]) != 0 {
		t.Fatalf("actor association should be empty slice")
	}
}
