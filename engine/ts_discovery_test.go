package engine

import (
	"os"
	"path/filepath"
	"strings"
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

// requireOnlyKeys checks that exactly the given keys are discovered, and that
// only those starting with "%d" are plural.
func requireOnlyKeys(t *testing.T, sources map[string]string, want ...Key) {
	t.Helper()
	keys, vars := discoverTestSourcesWithVars(t, sources)
	for _, key := range want {
		requireDiscoveredKey(t, keys, key, strings.HasPrefix(string(key), "%d"))
	}
	if len(keys) != len(want) || len(vars) != 0 {
		t.Fatalf("keys=%+v vars=%+v, want only %q", keys, vars, want)
	}
}

func TestTSXElementsDoNotHideCalls(t *testing.T) {
	requireOnlyKeys(t, map[string]string{"view.tsx": `
const a = <div><p>{t("A")}</p><p>{t("B")}</p></div>;
const b = <p><Icon name={x} />{t("C")}</p>;
const c = <>{t("D")}</>;
const d = <ul>
  <li class="x">{t("E")}</li>
  <Show when={open()} fallback={<span>{t("F")}</span>}>
    <input value={v} onInput={(e) => set(e.currentTarget.value)} />{t("G")}
  </Show>
</ul>;
`}, "A", "B", "C", "D", "E", "F", "G")
}

func TestTSXTextQuotesAreText(t *testing.T) {
	requireOnlyKeys(t, map[string]string{"view.tsx": `
const a = <p>Don't {t("Save")}</p>;
const b = <p>Serra's "map {t("Map")}</p>;
const c = <p title="t('attr') Don't" aria-label='t("attr2")'>{t("Title")}</p>;
const d = <p>{cond && <b>{t("N")}</b>} it's {t("After")}</p>;
`}, "Save", "Map", "Title", "N", "After")
}

func TestTSRegexesAndComparisons(t *testing.T) {
	sources := map[string]string{
		"code.ts": `
const a = x.replace(/a'b/, "") + t("R1");
if (/re"/.test(s)) t("R2");
function f() { return /'/.test(s) && t("R3"); }
const ratio = a / b / t("R4").length;
const less = a < b && c > d ? t("R5") : t("R6");
const list: Array<string> = [t("R7")];
const map = new Map<string, number>(); t("R8");
`,
		"generic.tsx": `
const id = <T,>(value: T) => value; t("G1");
const pick = <T extends object>(value: T) => value; t("G2");
const less = i < n ? t("G3") : <b>{t("G4")}</b>;
`,
	}
	requireOnlyKeys(t, sources, "R1", "R2", "R3", "R4", "R5", "R6", "R7", "R8", "G1", "G2", "G3", "G4")
}

func TestTSXRealWorldShapes(t *testing.T) {
	requireOnlyKeys(t, map[string]string{
		"components/EntryDialog.tsx": `import { createSignal, For, Show } from "solid-js";
import { t, tp } from "../i18n";
import { Icon } from "./Icon";

type Props = { entries: Entry[]; onClose: () => void };

export function EntryDialog(props: Props) {
  const [query, setQuery] = createSignal("");
  const visible = () => props.entries.filter((e) => e.name.toLowerCase().includes(query().toLowerCase()));
  return (
    <dialog class="entry-dialog" aria-label={t("Entries")} onKeyDown={(e) => e.key === "Escape" && props.onClose()}>
      <header>
        <h2>{t("Pick an entry")}</h2>
        <button type="button" title={t("Close")} onClick={props.onClose}>
          <Icon name="x" />
        </button>
      </header>
      {/* The list's "empty" state isn't a t("call") */}
      <input placeholder={t("Search…")} value={query()} onInput={(e) => setQuery(e.currentTarget.value)} />
      <Show when={visible().length > 0} fallback={<p class="empty">{t("No entry matches.")}</p>}>
        <ul>
          <For each={visible()}>{(entry) => <li classList={{ active: entry.id === current() }}>{entry.name}</li>}</For>
        </ul>
      </Show>
      <footer>{tp("%d entries", visible().length, visible().length)} — you'll see more soon</footer>
    </dialog>
  );
}
`,
		"lib/format.ts": `const placeholder = /%(\[\d+\])?[sd%]/g;
export function format(text: string, ...args: unknown[]): string {
  let next = 0;
  return text.replace(placeholder, (match) => (match === "%%" ? "%" : String(args[next++])));
}
export const deleteLabel = (name: string) => t("Delete %s?", name);
export const half = (n: number) => n / 2 < 1 ? t("Small") : t("Large");
`,
	}, "Entries", "Pick an entry", "Close", "Search…", "No entry matches.", "Delete %s?", "Small", "Large", "%d entries")
}
