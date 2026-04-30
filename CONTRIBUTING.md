# Contributing to mcp-withallo

Thanks for your interest. This is a small, focused MCP server that wraps the [Allo API](https://help.withallo.com/en/api-reference/introduction). Contributions are welcome — especially new read-only tools, sharper error messages, and docs.

## Prerequisites

- Go 1.25 or newer (`go version`)
- An Allo API key (Settings → API keys in the Allo dashboard) for end-to-end testing
- An MCP client (Claude Desktop, Claude Code, Cursor…) if you want to validate UX

## Build, run, test

```bash
git clone https://github.com/edouard-claude/mcp-withallo
cd mcp-withallo

make build         # or: go build -o mcp-withallo .
make lint          # gofmt + go vet — CI rejects unformatted code
make test          # go test -race ./...
make smoke-test    # full end-to-end with real API key (set ALLO_API_KEY first)
```

## Code layout

- `main.go` — entry point, env var, MCP server bootstrap
- `client.go` — minimal HTTP client (`Get`, `PostJSON`) wrapping the Allo API
- `tools.go` — every MCP tool definition + handler

That's it. No framework, no codegen.

## Adding a new tool

Each tool is a small `addXxx(s, c)` function in `tools.go`. The pattern:

1. Define the tool with `mcp.NewTool("allo_xxx", ...)` — name, description, params (with enums and defaults where it helps the model).
2. Register the handler with `s.AddTool(tool, func(ctx, req) (*mcp.CallToolResult, error) { ... })`.
3. In the handler: parse args, call `c.Get(...)` or `c.PostJSON(...)`, hand off to `respond(body, status, err)`.
4. Add the tool to `registerTools()` and the table in `README.md`.
5. Add a one-line entry to `CHANGELOG.md` under `[Unreleased]`.

Mirror an existing tool when in doubt: `addGetContact` is a clean GET-with-path-param example, `addAnalyticsOverview` is a clean POST-with-body example, `addSearchConversationItems` is the most complex with all filter types.

**Read-only invariant.** All shipped tools must be read-only. If you want to add a write operation (send SMS, mutate contact, CRM write), open an issue first to discuss scope and rate-limit implications — write endpoints share the same daily quota and need explicit user opt-in.

## Code style

- Run `gofmt -w .` (or `make lint` then `gofmt -w .`) before committing.
- Keep handlers small. If logic exceeds ~30 lines, extract a helper.
- Errors surfaced to the model must be a single actionable sentence — see `formatRateLimit` for the bar.

## Pull requests

- Branch from `main`.
- One logical change per PR. Update `README.md` and `CHANGELOG.md` if behavior changes.
- Describe how you tested it (smoke-test output, screenshot of Claude calling it, etc.).
- CI must be green (`gofmt`, `go vet`, `go build`, `go test -race`).

## Reporting bugs

Open an issue with the bug template. Include the MCP client (Claude Desktop / Claude Code / other), OS, Go version, and the redacted MCP server stderr.

For security issues, see [SECURITY.md](./SECURITY.md) — please **do not** open a public issue.
