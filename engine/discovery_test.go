package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/halsten-dev/lokyn"
)

func discoverTestSources(t *testing.T, sources map[string]string) DiscoveredKeys {
	t.Helper()
	keys, _ := discoverTestSourcesWithVars(t, sources)
	return keys
}

func discoverTestSourcesWithVars(t *testing.T, sources map[string]string) (DiscoveredKeys, DiscoveredVars) {
	t.Helper()
	lokyn.Init()
	lokyn.SetLanguage("en")
	root := t.TempDir()
	for name, content := range sources {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	keys, vars, err := DiscoverKeys(root)
	if err != nil {
		t.Fatal(err)
	}
	return keys, vars
}

func requireDiscoveredKey(t *testing.T, keys DiscoveredKeys, text Key, plural bool) {
	t.Helper()
	index := keys.KeyIndex(text)
	if index < 0 {
		t.Fatalf("missing key %q in %+v", text, keys)
	}
	if key := keys[index]; key.Err != nil || key.IsPlural != plural {
		t.Fatalf("key %q = %+v, want resolved, plural=%v", text, key, plural)
	}
}

func TestDiscoverKeysFollowsTwoWrapperLevels(t *testing.T) {
	keys := discoverTestSources(t, map[string]string{
		"wrappers.go": `package example
import tr "github.com/halsten-dev/lokyn"
func translate(key string) string { return tr.L(key) }
func outer(ignored int, message string) string { return translate(message) }
func plural(count int, message string) string { return tr.P(message, count) }
func outerPlural(message string, count int) string { return plural(count, message) }
func tooDeep(key string) string { return outer(0, key) }
`,
		"calls.go": `package example
var a = translate("One level")
var b = outer(1, "Two levels")
var c = plural(2, "Items")
var d = outerPlural("More items", 3)
var e = tooDeep("Beyond limit")
`,
	})
	requireDiscoveredKey(t, keys, "One level", false)
	requireDiscoveredKey(t, keys, "Two levels", false)
	requireDiscoveredKey(t, keys, "Items", true)
	requireDiscoveredKey(t, keys, "More items", true)
	if keys.ContainsKey("Beyond limit") || keys.ContainsKey("message") {
		t.Fatalf("unexpected wrapper keys: %+v", keys)
	}
}

func TestDiscoverKeysFollowsImportedWrappersAndMethods(t *testing.T) {
	keys := discoverTestSources(t, map[string]string{
		"go.mod": "module example.org/app\n\ngo 1.25\n",
		"text/wrappers.go": `package text
import "github.com/halsten-dev/lokyn"
func Translate(key string) string { return lokyn.L(key) }
type Translator struct{}
func (Translator) Items(count int, key string) string { return lokyn.P(key, count) }
`,
		"main.go": `package main
import words "example.org/app/text"
func local(key string) string { return words.Translate(key) }
var a = local("Imported")
var b = words.Translator{}.Items(2, "Method items")
var c = words.Translator.Items(words.Translator{}, 3, "Method expression items")
`,
	})
	requireDiscoveredKey(t, keys, "Imported", false)
	requireDiscoveredKey(t, keys, "Method items", true)
	requireDiscoveredKey(t, keys, "Method expression items", true)
	if len(keys) != 3 {
		t.Fatalf("unexpected keys: %+v", keys)
	}
}

func TestDiscoverKeysUsesGoSyntax(t *testing.T) {
	keys, vars := discoverTestSourcesWithVars(t, map[string]string{
		"main.go": "package example\n" +
			"import . \"github.com/halsten-dev/lokyn\"\n" +
			"// L(\"Comment\")\n" +
			"var text = `L(\"Not a call\")`\n" +
			"var a = L(\"Escaped \\\"quote\\\"\")\n" +
			"var b = L(\n `Raw key`,\n)\n" +
			"const prefix = \"Constant\"\n" +
			"var c = L(prefix + \" key\")\n" +
			"var d = P(\"Plural\", 2)\n" +
			"var dynamic string\nvar e = L(dynamic)\n",
	})
	requireDiscoveredKey(t, keys, "Escaped \"quote\"", false)
	requireDiscoveredKey(t, keys, "Raw key", false)
	requireDiscoveredKey(t, keys, "Constant key", false)
	requireDiscoveredKey(t, keys, "Plural", true)
	if len(vars) != 1 || vars[0].Expression != "dynamic" {
		t.Fatalf("dynamic key must be discovered separately: %+v", vars)
	}
	if len(keys) != 4 {
		t.Fatalf("unexpected keys: %+v", keys)
	}
}

func TestDiscoverKeysDoesNotConfuseSameNamedFunctions(t *testing.T) {
	keys := discoverTestSources(t, map[string]string{
		"a/a.go": `package a
import "github.com/halsten-dev/lokyn"
func translate(key string) string { return lokyn.L(key) }
var a = translate("Real")
`,
		"b/b.go": `package b
func translate(key string) string { return key }
var b = translate("Not translated")
`,
	})
	requireDiscoveredKey(t, keys, "Real", false)
	if len(keys) != 1 {
		t.Fatalf("unexpected keys: %+v", keys)
	}
}

func TestDiscoverKeysRequestLocalAPIWithNestedModule(t *testing.T) {
	keys := discoverTestSources(t, map[string]string{
		"src/go.mod": "module example.org/web\n\ngo 1.25\n",
		"src/util/lang.go": `package util
import (
 "github.com/gofiber/fiber/v2"
 "github.com/halsten-dev/lokyn"
)
func Localizer(c *fiber.Ctx) *lokyn.Localizer { return lokyn.NewLocalizer("en") }
func L(c *fiber.Ctx, key string) string { return Localizer(c).L(key) }
func P(c *fiber.Ctx, key string, count int) string { return Localizer(c).P(key, count) }
`,
		"src/play/util/lang.go": `package util
import (
 "github.com/gofiber/fiber/v2"
 web "example.org/web/util"
)
func L(c *fiber.Ctx, key string) string { return web.L(c, key) }
`,
		"src/main.go": `package main
import (
 "context"
 "github.com/halsten-dev/lokyn"
 "example.org/web/util"
 play "example.org/web/play/util"
)
var webKey = util.L(nil, "Web wrapper")
var playKey = play.L(nil, "Play wrapper")
var pluralKey = util.P(nil, "Plural wrapper", 2)
var ctx = context.Background()
var localizer = lokyn.NewLocalizer("en")
var a = localizer.L("Localizer key")
var b = localizer.P("Localizer plural", 3)
var c = lokyn.FromContext(ctx).L("From context")
var d = lokyn.LCtx(ctx, "Context key")
var e = lokyn.PCtx(ctx, "Context plural", 2)
var f = (*lokyn.Localizer).L(localizer, "Method expression")
func contextKey(ctx context.Context, key string) string { return lokyn.LCtx(ctx, key) }
func outer(key string) string { return contextKey(ctx, key) }
var g = outer("Context wrapper")
`,
	})
	for _, text := range []Key{"Web wrapper", "Play wrapper", "Localizer key", "From context", "Context key", "Method expression", "Context wrapper"} {
		requireDiscoveredKey(t, keys, text, false)
	}
	for _, text := range []Key{"Plural wrapper", "Localizer plural", "Context plural"} {
		requireDiscoveredKey(t, keys, text, true)
	}
	if len(keys) != 10 {
		t.Fatalf("unexpected keys: %+v", keys)
	}
}

func TestDiscoverKeysDoesNotTreatUnrelatedMethodsAsLokyn(t *testing.T) {
	keys := discoverTestSources(t, map[string]string{
		"main.go": `package example
import "github.com/halsten-dev/lokyn"
type Other struct{}
func (Other) L(key string) string { return key }
func (Other) P(key string, count int) string { return key }
var a = Other{}.L("Not Lokyn")
var b = Other{}.P("Not plural Lokyn", 2)
var c = lokyn.LCtx(nil, "Real context key")
`,
	})
	requireDiscoveredKey(t, keys, "Real context key", false)
	if len(keys) != 1 {
		t.Fatalf("unexpected keys: %+v", keys)
	}
}

func TestDiscoverKeysHandlesUnusedWrappersAndRecursion(t *testing.T) {
	keys, vars := discoverTestSourcesWithVars(t, map[string]string{
		"main.go": `package example
import "github.com/halsten-dev/lokyn"
func unused(key string) string { return lokyn.L(key) }
func recursive(message string) string {
 if false { return recursive(message) }
 return lokyn.L(message)
}
var a = recursive("Recursive")
`,
	})
	requireDiscoveredKey(t, keys, "Recursive", false)
	found := false
	for _, variable := range vars {
		if variable.Expression == "key" {
			found = true
		}
	}
	if !found {
		t.Fatalf("unused wrapper should retain manual reconciliation: %+v", vars)
	}
}
