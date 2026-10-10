package engine

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/constant"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/halsten-dev/lokyn"
)

const maxWrapperDepth = 2
const lokynImportPath = "github.com/halsten-dev/lokyn"

type sourcePackage struct {
	path  string
	files []*ast.File
	info  *types.Info
	pkg   *types.Package
}

type sourceCall struct {
	call  *ast.CallExpr
	owner *types.Func
	info  *types.Info
}

// sourceImporter resolves packages in the scanned tree. Other dependencies need
// not be installed or buildable: discovery only needs local function/parameter
// identities, not a successful type check of the entire application.
type sourceImporter struct {
	fset     *token.FileSet
	packages map[string]*sourcePackage
	checking map[string]bool
	external map[string]*types.Package
}

func (s *sourceImporter) Import(path string) (*types.Package, error) {
	if p := s.packages[path]; p != nil {
		if p.pkg != nil {
			return p.pkg, nil
		}
		if s.checking[path] {
			return nil, fmt.Errorf("import cycle: %s", path)
		}
		s.checking[path] = true
		config := types.Config{Importer: s, Error: func(error) {}}
		p.pkg, _ = config.Check(path, s.fset, p.files, p.info)
		delete(s.checking, path)
		return p.pkg, nil
	}
	if p := s.external[path]; p != nil {
		return p, nil
	}
	name := filepath.Base(path)
	if path == lokynImportPath {
		name = "lokyn"
	}
	p := types.NewPackage(path, name)
	if path == lokynImportPath {
		defineLokynAPI(p)
	}
	p.MarkComplete()
	s.external[path] = p
	return p, nil
}

// defineLokynAPI models both the process-wide API and the v1.1 request-local
// API without requiring the scanned project's dependencies to be installed.
func defineLokynAPI(p *types.Package) {
	localizer := types.NewNamed(types.NewTypeName(token.NoPos, p, "Localizer", nil), types.NewStruct(nil, nil), nil)
	p.Scope().Insert(localizer.Obj())
	pointer := types.NewPointer(localizer)
	// Context values are opaque to key discovery; an interface stand-in lets
	// us resolve translators even when the caller's context type is unavailable.
	context := types.NewInterfaceType(nil, nil).Complete()
	str, integer := types.Typ[types.String], types.Typ[types.Int]
	tuple := func(values []types.Type) *types.Tuple {
		vars := make([]*types.Var, len(values))
		for i, value := range values {
			vars[i] = types.NewParam(token.NoPos, p, "", value)
		}
		return types.NewTuple(vars...)
	}
	function := func(name string, receiver *types.Var, params []types.Type, result types.Type) *types.Func {
		sig := types.NewSignatureType(receiver, nil, nil, tuple(params), tuple([]types.Type{result}), false)
		return types.NewFunc(token.NoPos, p, name, sig)
	}
	for _, name := range []string{"L", "P"} {
		params := []types.Type{str}
		if name == "P" {
			params = append(params, integer)
		}
		p.Scope().Insert(function(name, nil, params, str))
		localizer.AddMethod(function(name, types.NewVar(token.NoPos, p, "lz", pointer), params, str))
		p.Scope().Insert(function(name+"Ctx", nil, append([]types.Type{context}, params...), str))
	}
	p.Scope().Insert(function("NewLocalizer", nil, []types.Type{str}, pointer))
	p.Scope().Insert(function("FromContext", nil, []types.Type{context}, pointer))
	p.Scope().Insert(function("WithContext", nil, []types.Type{context, pointer}, context))
}

func lokynKeyArgument(call sourceCall) (index int, plural bool, ok bool) {
	fn := calledFunction(call.call.Fun, call.info)
	if fn == nil || fn.Pkg() == nil || fn.Pkg().Path() != lokynImportPath {
		return 0, false, false
	}
	switch fn.Name() {
	case "L", "P":
		plural = fn.Name() == "P"
	case "LCtx", "PCtx":
		index = 1 // Context precedes the key.
		plural = fn.Name() == "PCtx"
	default:
		return 0, false, false
	}
	if sel, isSelector := call.call.Fun.(*ast.SelectorExpr); isSelector {
		if selection := call.info.Selections[sel]; selection != nil && selection.Kind() == types.MethodExpr {
			index++ // Explicit Localizer receiver precedes the key.
		}
	}
	return index, plural, true
}

// sourcePackagePath also works when discovery starts below the module root.
func sourcePackagePath(dir, name string) string {
	for parent := dir; ; parent = filepath.Dir(parent) {
		content, err := os.ReadFile(filepath.Join(parent, "go.mod"))
		if err == nil {
			for _, line := range strings.Split(string(content), "\n") {
				fields := strings.Fields(line)
				if len(fields) >= 2 && fields[0] == "module" {
					module := strings.Trim(fields[1], `"`)
					rel, _ := filepath.Rel(parent, dir)
					if rel != "." {
						module += "/" + filepath.ToSlash(rel)
					}
					if strings.HasSuffix(name, "_test") {
						module += "_test"
					}
					return module
				}
			}
		}
		if filepath.Dir(parent) == parent {
			return filepath.ToSlash(dir) + "/" + name
		}
	}
}

func discoverDirectory(keys *DiscoveredKeys, directory string) error {
	root, err := filepath.Abs(directory)
	if err != nil {
		return err
	}
	s := &sourceImporter{
		fset: token.NewFileSet(), packages: make(map[string]*sourcePackage),
		checking: make(map[string]bool), external: make(map[string]*types.Package),
	}
	var ordered []*sourcePackage
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path != root && strings.HasPrefix(entry.Name(), ".") {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}
		file, err := parser.ParseFile(s.fset, path, nil, 0)
		if err != nil {
			return err
		}
		pkgPath := sourcePackagePath(filepath.Dir(path), file.Name.Name)
		p := s.packages[pkgPath]
		if p == nil {
			p = &sourcePackage{path: pkgPath, info: &types.Info{
				Defs: make(map[*ast.Ident]types.Object), Uses: make(map[*ast.Ident]types.Object),
				Types: make(map[ast.Expr]types.TypeAndValue), Selections: make(map[*ast.SelectorExpr]*types.Selection),
			}}
			s.packages[pkgPath] = p
			ordered = append(ordered, p)
		}
		p.files = append(p.files, file)
		return nil
	})
	if err != nil {
		return err
	}
	for _, p := range ordered {
		if _, err := s.Import(p.path); err != nil {
			return err
		}
	}

	var direct []sourceCall
	callers := make(map[*types.Func][]sourceCall)
	for _, p := range ordered {
		for _, file := range p.files {
			// Walk each declaration separately to retain the owning function.
			for _, decl := range file.Decls {
				var owner *types.Func
				if fn, ok := decl.(*ast.FuncDecl); ok {
					owner, _ = p.info.Defs[fn.Name].(*types.Func)
				}
				ast.Inspect(decl, func(node ast.Node) bool {
					call, ok := node.(*ast.CallExpr)
					if !ok {
						return true
					}
					fn := calledFunction(call.Fun, p.info)
					if fn == nil {
						return true
					}
					c := sourceCall{call: call, owner: owner, info: p.info}
					callers[fn] = append(callers[fn], c)
					if _, _, ok := lokynKeyArgument(c); ok {
						direct = append(direct, c)
					}
					return true
				})
			}
		}
	}

	var trace func(ast.Expr, sourceCall, bool, int)
	trace = func(expr ast.Expr, context sourceCall, plural bool, depth int) {
		if value := context.info.Types[expr].Value; value != nil && value.Kind() == constant.String {
			appendDiscoveredKey(keys, constant.StringVal(value), plural, nil)
			return
		}
		if index := parameterIndex(expr, context); index >= 0 && depth < maxWrapperDepth {
			followed := false
			for _, caller := range callers[context.owner] {
				argument := index
				// A method expression (T.Translate(receiver, key)) includes
				// the receiver as an explicit argument, unlike receiver.Translate(key).
				if sel, ok := caller.call.Fun.(*ast.SelectorExpr); ok {
					if selection := caller.info.Selections[sel]; selection != nil && selection.Kind() == types.MethodExpr {
						argument++
					}
				}
				if argument < len(caller.call.Args) {
					trace(caller.call.Args[argument], caller, plural, depth+1)
					followed = true
				}
			}
			if followed {
				return
			}
		}
		var text bytes.Buffer
		_ = format.Node(&text, s.fset, expr)
		appendDiscoveredKey(keys, text.String(), plural, errors.New(lokyn.L("Key is variable, need manual matching")))
	}
	for _, call := range direct {
		index, plural, _ := lokynKeyArgument(call)
		if index < len(call.call.Args) {
			trace(call.call.Args[index], call, plural, 0)
		}
	}
	return nil
}

func calledFunction(expr ast.Expr, info *types.Info) *types.Func {
	switch e := expr.(type) {
	case *ast.Ident:
		fn, _ := info.Uses[e].(*types.Func)
		return fn
	case *ast.SelectorExpr:
		fn, _ := info.Uses[e.Sel].(*types.Func)
		return fn
	case *ast.ParenExpr:
		return calledFunction(e.X, info)
	case *ast.IndexExpr:
		return calledFunction(e.X, info)
	case *ast.IndexListExpr:
		return calledFunction(e.X, info)
	}
	return nil
}

func parameterIndex(expr ast.Expr, context sourceCall) int {
	if parens, ok := expr.(*ast.ParenExpr); ok {
		return parameterIndex(parens.X, context)
	}
	id, ok := expr.(*ast.Ident)
	if !ok || context.owner == nil {
		return -1
	}
	sig, ok := context.owner.Type().(*types.Signature)
	if !ok {
		return -1
	}
	for i := 0; i < sig.Params().Len(); i++ {
		if context.info.Uses[id] == sig.Params().At(i) {
			return i
		}
	}
	return -1
}

func appendDiscoveredKey(keys *DiscoveredKeys, text string, plural bool, err error) {
	key := Key(text)
	if !keys.ContainsKey(key) {
		*keys = append(*keys, DiscoveredKey{Key: key, IsPlural: plural, Err: err})
	}
}
