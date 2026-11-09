package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTSDiscoversLiteralKeys(t *testing.T) {
	keys, vars := discoverTestSourcesWithVars(t, map[string]string{
		"src/view.tsx": "import { t, tKey, tp } from \"./i18n\";\n" +
			"const a = t(\"Double\");\n" +
			"const b = t( 'Single' );\n" +
			"const c = t(`Backtick`);\n" +
			"const d = t(\"Escaped \\\"quote\\\"\");\n" +
			"const e = t('It\\'s');\n" +
			"const f = t(\"Delete %s?\", name);\n" +
			"const g = tKey(\"Marked\");\n" +
			"const h = tp(\"%d entries\", count, count);\n" +
			"const i = <p title={`${t(\"In template\")}`}>{t(\n  \"Multiline\",\n)}</p>;\n",
		"src/plain.ts": "export const x = t(\"Plain ts\");\n",
	})
	for _, key := range []Key{"Double", "Single", "Backtick", `Escaped "quote"`, "It's", "Delete %s?", "Marked", "In template", "Multiline", "Plain ts"} {
		requireDiscoveredKey(t, keys, key, false)
	}
	requireDiscoveredKey(t, keys, "%d entries", true)
	if len(keys) != 11 || len(vars) != 0 {
		t.Fatalf("keys=%+v vars=%+v", keys, vars)
	}
}

func TestTSUnescapesLiterals(t *testing.T) {
	for raw, want := range map[string]string{
		`plain`:             "plain",
		`a\nb\tc`:           "a\nb\tc",
		`\x41\u0042\u{43}!`: "ABC!",
		`\u{1F600}`:         "\U0001F600",
		"line\\\nbreak":     "linebreak",
		`\q\\`:              `q\`,
	} {
		if got := unescapeTypeScript([]byte(raw)); got != want {
			t.Errorf("unescape(%q) = %q, want %q", raw, got, want)
		}
	}
}

func TestTSIgnoresCommentsStringsAndMembers(t *testing.T) {
	keys, vars := discoverTestSourcesWithVars(t, map[string]string{
		"src/view.ts": "// t(\"x\")\n" +
			"/* t(\"block\") */\n" +
			"const s = \"t('y')\";\n" +
			"const u = `t(\"in template\")`;\n" +
			"const r = /[\"']t(\"regex\")/;\n" +
			"obj.t(\"z\");\n" +
			"obj?.t(\"optional\");\n" +
			"format(\"w\");\n" +
			"const t2 = 1;\n" +
			"export function t(key: string): string { return key; }\n" +
			"const after = t(\"Real\");\n",
	})
	requireDiscoveredKey(t, keys, "Real", false)
	if len(keys) != 1 || len(vars) != 0 {
		t.Fatalf("keys=%+v vars=%+v", keys, vars)
	}
}

func TestTSPluralAndSingularConflict(t *testing.T) {
	keys := discoverTestSources(t, map[string]string{
		"a.ts": "t(\"Same\");\n",
		"b.ts": "tp(\"Same\", 2);\n",
	})
	if len(keys) != 1 || keys[0].Err == nil {
		t.Fatalf("missing conflict: %+v", keys)
	}
}

func TestTSDynamicKeyIsAVariable(t *testing.T) {
	keys, vars := discoverTestSourcesWithVars(t, map[string]string{
		"src/list.tsx": "\nconst label = t(labels[id]);\n" +
			"const again = t( labels[id] );\n" +
			"const count = tp(`${n} items`, n);\n",
	})
	if len(keys) != 0 || len(vars) != 2 {
		t.Fatalf("keys=%+v vars=%+v", keys, vars)
	}
	want := DiscoveredVar{Var: "src/list.tsx:labels[id]", Expression: "labels[id]", File: "src/list.tsx", Line: 2}
	if got := vars[0]; got != want {
		t.Fatalf("var = %+v, want %+v", got, want)
	}
	if got := vars[1]; got.Expression != "`${n} items`" || !got.IsPlural || got.Line != 4 {
		t.Fatalf("template var = %+v", got)
	}
}

func TestTSSkipsNodeModulesAndHidden(t *testing.T) {
	keys := discoverTestSources(t, map[string]string{
		"node_modules/pkg/index.ts": "t(\"Dependency\");\n",
		".cache/view.ts":            "t(\"Hidden\");\n",
		"dist/view.ts":              "t(\"Built\");\n",
		"bin/view.ts":               "t(\"Binary\");\n",
		"bindings/api.ts":           "t(\"Binding\");\n",
		"src/view.ts":               "t(\"Kept\");\n",
	})
	requireDiscoveredKey(t, keys, "Kept", false)
	if len(keys) != 1 {
		t.Fatalf("unexpected keys: %+v", keys)
	}
}

func TestGoDiscoversLokynKey(t *testing.T) {
	keys, vars := discoverTestSourcesWithVars(t, map[string]string{
		"main.go": `package example
import "github.com/halsten-dev/lokyn"
func mark(key string) string { return lokyn.Key(key) }
var a = lokyn.Key("Direct")
var b = mark("Wrapped")
`,
	})
	requireDiscoveredKey(t, keys, "Direct", false)
	requireDiscoveredKey(t, keys, "Wrapped", false)
	if len(keys) != 2 || len(vars) != 0 {
		t.Fatalf("keys=%+v vars=%+v", keys, vars)
	}
}

func TestRelativeLocationPath(t *testing.T) {
	for _, saved := range []string{"", "."} {
		t.Run(saved, func(t *testing.T) {
			root := t.TempDir()
			project := ProjectNew(root, "translations", []string{"en"})
			project.LocationPath = saved
			configPath := filepath.Join(root, PROJECT_DIR_NAME, PROJECT_FILE_NAME)
			if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
				t.Fatal(err)
			}
			if err := projectSave(&project, configPath); err != nil {
				t.Fatal(err)
			}
			before, err := os.ReadFile(configPath)
			if err != nil {
				t.Fatal(err)
			}

			discovered, err := DiscoverProject(root)
			if err != nil {
				t.Fatal(err)
			}
			if discovered.LocationPath != root {
				t.Fatalf("LocationPath = %q, want %q", discovered.LocationPath, root)
			}
			// The committed project file must stay portable across checkouts.
			if err := ProjectSave(&discovered); err != nil {
				t.Fatal(err)
			}
			after, err := os.ReadFile(configPath)
			if err != nil {
				t.Fatal(err)
			}
			if string(after) != string(before) {
				t.Fatalf("project file changed:\n%s\nwant:\n%s", after, before)
			}
		})
	}
}
