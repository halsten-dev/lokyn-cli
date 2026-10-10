# lokyn-cli
Lokyn CLI is a tool to help translate a multilingual project created with the lib Lokyn.

## Discovery

Go source discovery recognizes Lokyn `L`, `P`, `LCtx`, and `PCtx`, including
import aliases and `Localizer.L` / `Localizer.P` methods. It follows forwarded
key parameters through up to **two wrapper-function levels**, across files and
local packages. String literals and string constants are resolved automatically.
Hidden files and folders are skipped; generated Go files and tests are included.

## Dynamic-key reconciliation

Dynamic expressions are not translation keys. For example:

```go
func render(status string) {
    lokyn.L(status)
}
```

Link `status` to concrete runtime keys such as `"pending"`, `"approved"`, and
`"rejected"`. The concrete keys join the translation/export workflow; the source
expression itself is never exported. Links do not rewrite application code.

The reconciliation screen contains new keys, unused translations, and dynamic
expressions. Use **Tab / Shift+Tab** to change lists.

| Key | Action |
| --- | --- |
| `e` on an expression | Open its linked-key editor |
| `i` on an expression | Explicitly ignore it (saved links are kept but inactive) |
| `r` on an expression | Reset its links and ignore decision |
| `a` | Add a manual concrete key; edit it with `e` |
| `d` on a key | Remove a new/manual or unused key |
| `/` | Filter the focused list |
| `Enter` | Validate and save associations, then continue to translations |

In the linked-key editor:

- **Space** links the selected available key.
- **Tab** switches between available and linked keys.
- **d** removes the selected linked key.
- **a** enters a new concrete key; **Enter** adds it, **Esc** cancels input.
- **Enter** saves the link selection; **Esc** cancels the dialog.
- At least one linked key is required. To exclude an expression, use `i` instead.

Every discovered expression must have links or an explicit ignore decision
before validation succeeds. Linked keys are protected from unused-key removal.
Existing translation values are preserved, and new keys get entries for every
managed language. Empty/duplicate keys and singular/plural conflicts are rejected.
A failed project save leaves you on reconciliation without replacing translations.

Associations and ignore decisions are stored in `.lokyn/lokynproj.json`.
Expression IDs include relative source paths and lexical scope, not line numbers,
so identical names in different functions/blocks remain independent. Moving the
project or inserting blank lines does not invalidate links. Renaming functions,
moving source files, or changing lexical structure may require relinking.

Saved associations for expressions no longer discovered remain stored for review
but do not activate translation keys. Their count is reported on entry. Review
or remove those entries in the project JSON when appropriate. Legacy mappings
keyed only by expression text are applied automatically only when unambiguous.
