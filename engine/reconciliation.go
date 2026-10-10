package engine

import (
	"fmt"
	"slices"
	"strings"
)

func (p *Project) InitVariableLinks() {
	if p.VarsKeysLink == nil {
		p.VarsKeysLink = make(map[Var]Keys)
	}
	if p.VarsIgnored == nil {
		p.VarsIgnored = make(map[Var]bool)
	}
}

// CloneVariableLinks keeps unvalidated UI edits separate from the loaded project.
func (p Project) CloneVariableLinks() Project {
	copy := p
	copy.VarsKeysLink = make(map[Var]Keys, len(p.VarsKeysLink))
	copy.VarsIgnored = make(map[Var]bool, len(p.VarsIgnored))
	for variable, keys := range p.VarsKeysLink {
		copy.VarsKeysLink[variable] = slices.Clone(keys)
	}
	for variable, ignored := range p.VarsIgnored {
		copy.VarsIgnored[variable] = ignored
	}
	return copy
}

func (p Project) VariableStatus(variable Var) VariableStatus {
	if p.VarsIgnored[variable] {
		return VariableIgnored
	}
	if len(p.VarsKeysLink[variable]) > 0 {
		return VariableLinked
	}
	return VariableUnresolved
}

func ValidateLinkedKeys(keys Keys) error {
	if len(keys) == 0 {
		return fmt.Errorf("link at least one concrete key, or explicitly ignore the expression")
	}
	seen := make(map[Key]bool)
	for _, key := range keys {
		if strings.TrimSpace(string(key)) == "" {
			return fmt.Errorf("translation keys cannot be empty")
		}
		if seen[key] {
			return fmt.Errorf("duplicate linked key %q", key)
		}
		seen[key] = true
	}
	return nil
}

func (p *Project) LinkVariable(variable Var, keys Keys) error {
	if err := ValidateLinkedKeys(keys); err != nil {
		return err
	}
	p.InitVariableLinks()
	p.VarsKeysLink[variable] = slices.Clone(keys)
	delete(p.VarsIgnored, variable)
	return nil
}

func (p *Project) IgnoreVariable(variable Var) {
	p.InitVariableLinks()
	// Keep links for review/reuse; ignored links do not activate keys.
	p.VarsIgnored[variable] = true
}

func (p *Project) ResetVariable(variable Var) {
	delete(p.VarsKeysLink, variable)
	delete(p.VarsIgnored, variable)
}

// MigrateLegacyLinks only transfers expression-name mappings when their scope
// is unambiguous. Ambiguous and disappeared entries remain stored, never applied.
func (p *Project) MigrateLegacyLinks(vars DiscoveredVars) {
	p.InitVariableLinks()
	counts := make(map[string]int)
	for _, variable := range vars {
		counts[variable.Expression]++
	}
	for _, variable := range vars {
		if counts[variable.Expression] != 1 || p.VariableStatus(variable.Var) != VariableUnresolved {
			continue
		}
		legacy := Var(variable.Expression)
		if keys, ok := p.VarsKeysLink[legacy]; ok {
			p.VarsKeysLink[variable.Var] = slices.Clone(keys)
		}
		if p.VarsIgnored[legacy] {
			p.VarsIgnored[variable.Var] = true
		}
	}
}

// LinkedDiscoveredKeys returns keys from active expressions only. Stale saved
// links remain on disk but must not keep unrelated translation keys alive.
func (p Project) LinkedDiscoveredKeys(vars DiscoveredVars) DiscoveredKeys {
	var keys DiscoveredKeys
	for _, variable := range vars {
		if p.VariableStatus(variable.Var) != VariableLinked {
			continue
		}
		for _, key := range p.VarsKeysLink[variable.Var] {
			appendDiscoveredKey(&keys, string(key), variable.IsPlural, variable.Err)
		}
	}
	return keys
}

func (p Project) DormantVariableCount(vars DiscoveredVars) int {
	seen := make(map[Var]bool)
	for variable := range p.VarsKeysLink {
		seen[variable] = true
	}
	for variable, ignored := range p.VarsIgnored {
		if ignored {
			seen[variable] = true
		}
	}
	for _, variable := range vars {
		delete(seen, variable.Var)
	}
	return len(seen)
}

// ReconcileTranslations validates the entire selection before constructing a
// fresh map. It never mutates existing translations or the saved project.
func ReconcileTranslations(project Project, vars DiscoveredVars, selected DiscoveredKeys, existing KeyLangMap) (KeyLangMap, error) {
	if len(project.ManagedLanguages) == 0 {
		return nil, fmt.Errorf("project must have at least one managed language")
	}
	modes := make(map[Key]bool)
	add := func(key Key, plural bool, err error) error {
		if strings.TrimSpace(string(key)) == "" {
			return fmt.Errorf("translation keys cannot be empty")
		}
		if err != nil {
			return fmt.Errorf("key %q: %w", key, err)
		}
		if previous, ok := modes[key]; ok && previous != plural {
			return fmt.Errorf("key %q is required as both singular and plural", key)
		}
		modes[key] = plural
		return nil
	}
	for _, key := range selected {
		if _, ok := modes[key.Key]; ok {
			return nil, fmt.Errorf("duplicate selected key %q", key.Key)
		}
		if err := add(key.Key, key.IsPlural, key.Err); err != nil {
			return nil, err
		}
	}
	for _, variable := range vars {
		if project.VariableStatus(variable.Var) == VariableIgnored {
			continue
		}
		if variable.Err != nil {
			return nil, fmt.Errorf("%s (%s): %w", variable.Expression, variable.File, variable.Err)
		}
		if err := ValidateLinkedKeys(project.VarsKeysLink[variable.Var]); err != nil {
			return nil, fmt.Errorf("%s (%s, %s): %w", variable.Expression, variable.File, variable.Scope, err)
		}
		for _, key := range project.VarsKeysLink[variable.Var] {
			if err := add(key, variable.IsPlural, nil); err != nil {
				return nil, err
			}
		}
	}
	result := make(KeyLangMap, len(modes))
	for key, plural := range modes {
		result[key] = make(map[Lang]Translation)
		// Keep all imported languages/values; fill missing managed languages below.
		for lang, translation := range existing[key] {
			if translation.IsPlural != plural && (translation.OneValue != "" || translation.OtherValue != "") {
				return nil, fmt.Errorf("key %q has an existing %s translation with a conflicting plural type", key, lang)
			}
			translation.Key, translation.Lang, translation.IsPlural = key, lang, plural
			result[key][lang] = translation
		}
		for _, lang := range project.ManagedLanguages {
			if _, ok := result[key][lang]; !ok {
				result[key][lang] = Translation{Key: key, Lang: lang, IsPlural: plural}
			}
		}
	}
	return result, nil
}
