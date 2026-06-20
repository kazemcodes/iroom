# Project Rules & Agent Context

## Tool Usage & Code Editing
* When using the `edit` tool on files (such as `cmd/server/main.go`), you **must** always supply the `oldString` argument containing the exact, literal block of code you intend to replace.
* Do not omit `oldString` or attempt partial/malformed JSON schemas when issuing search-and-replace instructions.
* If a schema validation error occurs regarding a missing key, instantly fall back to printing the full updated code block in standard Markdown or request manual review.