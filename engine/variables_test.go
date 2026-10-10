package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDynamicExpressionIdentityAndPluralUsage(t *testing.T) {
	source := `package example
import "github.com/halsten-dev/lokyn"
func first(key string) {
 lokyn.L(key)
 if true { lokyn.L(key) }
 { key := "shadowed"; lokyn.L(key) }
}
func second(key string) { lokyn.P(key, 2) }
`
	keys, vars := discoverTestSourcesWithVars(t, map[string]string{"nested/main.go": source})
	if len(keys) != 0 || len(vars) != 3 {
		t.Fatalf("keys=%+v vars=%+v", keys, vars)
	}
	ids := make(map[Var]bool)
	plural := 0
	for _, variable := range vars {
		if variable.File != "nested/main.go" || variable.Expression != "key" || variable.Line == 0 || variable.Scope == "" {
			t.Fatalf("missing source context: %+v", variable)
		}
		if ids[variable.Var] {
			t.Fatalf("scope collision: %+v", vars)
		}
		ids[variable.Var] = true
		if variable.IsPlural {
			plural++
		}
	}
	if plural != 1 {
		t.Fatalf("plural count=%d", plural)
	}
	_, moved := discoverTestSourcesWithVars(t, map[string]string{"nested/main.go": "\n\n" + source})
	for _, variable := range moved {
		if !ids[variable.Var] {
			t.Fatalf("identity changed with whitespace/root: %+v", variable)
		}
	}
}

func TestDiscoveryReportsSingularPluralConflicts(t *testing.T) {
	keys, vars := discoverTestSourcesWithVars(t, map[string]string{"main.go": `package example
import "github.com/halsten-dev/lokyn"
func render(key string) { lokyn.L(key); lokyn.P(key, 2) }
var a = lokyn.L("Same")
var b = lokyn.P("Same", 2)
`})
	if len(keys) != 1 || keys[0].Err == nil || len(vars) != 1 || vars[0].Err == nil {
		t.Fatalf("missing conflicts: keys=%+v vars=%+v", keys, vars)
	}
}

func TestReconcileVariablesPreservesTranslationsAndFillsLanguages(t *testing.T) {
	project := ProjectNew(t.TempDir(), ".", []string{"en", "fr"})
	vars := DiscoveredVars{{Var: "status-id", Expression: "status", File: "main.go", Scope: "render"}, {Var: "items-id", Expression: "items", IsPlural: true}}
	if err := project.LinkVariable("status-id", Keys{"pending", "approved"}); err != nil {
		t.Fatal(err)
	}
	if err := project.LinkVariable("items-id", Keys{"items"}); err != nil {
		t.Fatal(err)
	}
	project.VarsKeysLink["disappeared"] = Keys{"obsolete"}
	existing := KeyLangMap{"pending": {"en": {OneValue: "Waiting"}}, "obsolete": {"en": {OneValue: "Old"}}}
	before := existing["pending"]["en"]
	data, err := ReconcileTranslations(project, vars, DiscoveredKeys{{Key: "literal"}}, existing)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 4 || data["pending"]["en"].OneValue != "Waiting" || data["pending"]["fr"].Lang != "fr" || !data["items"]["fr"].IsPlural || data["items"]["fr"].Key != "items" {
		t.Fatalf("bad reconciliation: %+v", data)
	}
	if existing["pending"]["en"] != before || len(existing) != 2 {
		t.Fatal("mutated existing data")
	}
	if project.DormantVariableCount(vars) != 1 || len(project.VarsKeysLink["disappeared"]) != 1 {
		t.Fatal("stale mappings were removed")
	}
}

func TestReconcileValidation(t *testing.T) {
	for _, test := range []struct {
		name    string
		prepare func(*Project, *DiscoveredVars, *DiscoveredKeys, KeyLangMap)
	}{
		{"unresolved", func(p *Project, v *DiscoveredVars, k *DiscoveredKeys, d KeyLangMap) { p.ResetVariable("id") }},
		{"empty linked key", func(p *Project, v *DiscoveredVars, k *DiscoveredKeys, d KeyLangMap) { p.VarsKeysLink["id"] = Keys{" "} }},
		{"duplicate linked key", func(p *Project, v *DiscoveredVars, k *DiscoveredKeys, d KeyLangMap) {
			p.VarsKeysLink["id"] = Keys{"value", "value"}
		}},
		{"empty manual key", func(p *Project, v *DiscoveredVars, k *DiscoveredKeys, d KeyLangMap) { *k = DiscoveredKeys{{Key: ""}} }},
		{"duplicate selected", func(p *Project, v *DiscoveredVars, k *DiscoveredKeys, d KeyLangMap) {
			*k = DiscoveredKeys{{Key: "x"}, {Key: "x"}}
		}},
		{"source plural conflict", func(p *Project, v *DiscoveredVars, k *DiscoveredKeys, d KeyLangMap) {
			(*v)[0].IsPlural = true
			*k = DiscoveredKeys{{Key: "value"}}
		}},
		{"imported plural conflict", func(p *Project, v *DiscoveredVars, k *DiscoveredKeys, d KeyLangMap) {
			d["value"] = map[Lang]Translation{"en": {IsPlural: true, OneValue: "One", OtherValue: "Many"}}
		}},
		{"variable error", func(p *Project, v *DiscoveredVars, k *DiscoveredKeys, d KeyLangMap) {
			(*v)[0].Err = fmt.Errorf("conflict")
		}},
		{"no languages", func(p *Project, v *DiscoveredVars, k *DiscoveredKeys, d KeyLangMap) { p.ManagedLanguages = nil }},
	} {
		t.Run(test.name, func(t *testing.T) {
			project := ProjectNew(t.TempDir(), ".", []string{"en"})
			project.LinkVariable("id", Keys{"value"})
			vars := DiscoveredVars{{Var: "id", Expression: "key"}}
			var selected DiscoveredKeys
			data := make(KeyLangMap)
			test.prepare(&project, &vars, &selected, data)
			if _, err := ReconcileTranslations(project, vars, selected, data); err == nil {
				t.Fatal("expected validation failure")
			}
		})
	}
}

func TestIgnoreResetAndSharedLinkedKeys(t *testing.T) {
	project := ProjectNew(t.TempDir(), ".", []string{"en"})
	vars := DiscoveredVars{{Var: "a", Expression: "a"}, {Var: "b", Expression: "b"}}
	project.LinkVariable("a", Keys{"shared"})
	project.LinkVariable("b", Keys{"shared"})
	data, err := ReconcileTranslations(project, vars, nil, nil)
	if err != nil || len(data) != 1 {
		t.Fatalf("shared link data=%+v err=%v", data, err)
	}
	project.IgnoreVariable("a")
	project.IgnoreVariable("b")
	if len(project.LinkedDiscoveredKeys(vars)) != 0 || project.VariableStatus("a") != VariableIgnored {
		t.Fatal("ignored links must not activate keys")
	}
	if data, err := ReconcileTranslations(project, vars, nil, nil); err != nil || len(data) != 0 {
		t.Fatalf("ignored data=%+v err=%v", data, err)
	}
	project.ResetVariable("b")
	if project.VariableStatus("b") != VariableUnresolved {
		t.Fatal("reset must require a new decision")
	}
	if _, err := ReconcileTranslations(project, vars, nil, nil); err == nil {
		t.Fatal("reset variable must block validation")
	}
}

func TestVariableMappingsPersistAndLegacyProjectsLoad(t *testing.T) {
	root := t.TempDir()
	project := ProjectNew(root, ".", []string{"en"})
	project.LinkVariable("id", Keys{"one", "two"})
	project.IgnoreVariable("ignored")
	if err := ProjectSave(&project); err != nil {
		t.Fatal(err)
	}
	loaded, err := DiscoverProject(root)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(loaded.VarsKeysLink, project.VarsKeysLink) || !loaded.VarsIgnored["ignored"] {
		t.Fatalf("lost links: %+v", loaded)
	}
	copy := loaded.CloneVariableLinks()
	copy.VarsKeysLink["id"][0] = "changed"
	copy.IgnoreVariable("new")
	if loaded.VarsKeysLink["id"][0] != "one" || loaded.VarsIgnored["new"] {
		t.Fatal("draft changes leaked")
	}
	legacy := filepath.Join(root, "legacy.json")
	if err := os.WriteFile(legacy, []byte(`{"ManagedLanguages":["en"]}`), 0644); err != nil {
		t.Fatal(err)
	}
	var old Project
	if err := projectLoad(&old, legacy); err != nil {
		t.Fatal(err)
	}
	if old.VarsKeysLink == nil || old.VarsIgnored == nil {
		t.Fatal("legacy maps not initialized")
	}
}

func TestProjectSaveFailureCleansTemporaryFile(t *testing.T) {
	root := t.TempDir()
	project := ProjectNew(root, ".", []string{"en"})
	// A directory cannot be replaced by the temporary project file.
	target := filepath.Join(root, "existing")
	if err := os.Mkdir(target, 0755); err != nil {
		t.Fatal(err)
	}
	sentinel := filepath.Join(target, "sentinel")
	if err := os.WriteFile(sentinel, []byte("unchanged"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := projectSave(&project, target); err == nil {
		t.Fatal("expected failed replacement")
	}
	content, err := os.ReadFile(sentinel)
	if err != nil || string(content) != "unchanged" {
		t.Fatalf("previous file damaged: %q err=%v", content, err)
	}
	leftovers, err := filepath.Glob(filepath.Join(root, ".lokyn-project-*"))
	if err != nil || len(leftovers) != 0 {
		t.Fatalf("temporary files left behind: %v err=%v", leftovers, err)
	}
}

func TestLegacyLinkMigrationIsScopeSafe(t *testing.T) {
	project := Project{}
	project.InitVariableLinks()
	project.VarsKeysLink["key"] = Keys{"old"}
	project.VarsKeysLink["unique"] = Keys{"unique-key"}
	vars := DiscoveredVars{{Var: "a", Expression: "key"}, {Var: "b", Expression: "key"}, {Var: "c", Expression: "unique"}}
	project.MigrateLegacyLinks(vars)
	if project.VariableStatus("a") != VariableUnresolved || project.VariableStatus("b") != VariableUnresolved || project.VariableStatus("c") != VariableLinked {
		t.Fatalf("unsafe migration: %+v", project.VarsKeysLink)
	}
}
