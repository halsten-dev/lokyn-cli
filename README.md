# lokyn-cli
Lokyn CLI is a tool to help translate a multilingual project created with the lib Lokyn.

## Key discovery

Discovery scans Go files for Lokyn `.L`, `.P`, `.LCtx`, and `.PCtx` calls,
including import aliases and `Localizer.L` / `Localizer.P` methods.
It follows forwarded key parameters through up to **two wrapper-function levels**
across files and local packages (including methods):

```go
func translate(key string) string { return lokyn.L(key) }
func message(key string) string { return translate(key) }

message("Hello") // Discovers "Hello".
```

Plural wrappers retain the plural flag. String literals and string constants are
resolved automatically; dynamic expressions, transformed wrapper parameters, and
calls beyond the two-level limit require manual matching. Hidden files and
folders are skipped.
