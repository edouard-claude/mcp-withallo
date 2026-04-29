# mcp-withallo

A local MCP server (stdio transport) that exposes the [Allo API](https://help.withallo.com/en/api-reference/introduction) as **read-only** tools to Claude Desktop and Claude Code.

## Tools (12)

All tools are read-only — no SMS sending, no contact mutations, no CRM writes.

| Tool | Allo endpoint |
|---|---|
| `allo_me` | `GET /v2/api/me` |
| `allo_list_numbers` | `GET /v2/api/numbers` |
| `allo_list_users` | `GET /v2/api/users` |
| `allo_list_tags` | `GET /v2/api/tags` |
| `allo_search_calls` *(includes transcripts)* | `GET /v1/api/calls` |
| `allo_search_contacts` | `GET /v1/api/contacts` |
| `allo_get_contact` | `GET /v1/api/contact/{id}` |
| `allo_get_contact_conversation` | `GET /v1/api/contact/{id}/conversation` |
| `allo_list_conversations` | `GET /v2/api/conversations` |
| `allo_search_conversation_items` | `POST /v2/api/conversations/items/search` |
| `allo_get_conversation_item` | `GET /v2/api/conversations/items/{id}` |
| `allo_analytics_overview` | `POST /v2/api/analytics/overview` |

## Install

### Option A — `go install` (recommended)

```bash
go install github.com/edouard-claude/mcp-withallo@latest
```

The binary is placed at `$GOBIN/mcp-withallo` (or `$GOPATH/bin/mcp-withallo`, typically `~/go/bin/mcp-withallo`). Make sure that directory is on your `PATH` if you want to invoke it without a full path.

### Option B — build from source

```bash
git clone https://github.com/edouard-claude/mcp-withallo
cd mcp-withallo
go build -o mcp-withallo .
```

The binary is produced in the current directory.

## Configuration

Get your API key from the Allo dashboard (Settings → API keys) and put it in `ALLO_API_KEY`. The key is sent verbatim in the `Authorization` header (no `Bearer` prefix).

### Claude Desktop

Edit `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "allo": {
      "command": "mcp-withallo",
      "env": { "ALLO_API_KEY": "your-api-key-here" }
    }
  }
}
```

Fully quit Claude (⌘Q) and relaunch.

### Claude Code

Option 1 — global, in `~/.claude/settings.json`:

```json
{
  "mcpServers": {
    "allo": {
      "command": "mcp-withallo",
      "env": { "ALLO_API_KEY": "your-api-key-here" }
    }
  }
}
```

Option 2 — per project, in a `.mcp.json` file at the repo root:

```json
{
  "mcpServers": {
    "allo": {
      "command": "mcp-withallo",
      "env": { "ALLO_API_KEY": "your-api-key-here" }
    }
  }
}
```

If `mcp-withallo` is not on your `PATH`, replace `"command": "mcp-withallo"` with the absolute path (run `which mcp-withallo` to find it — typically `$(go env GOPATH)/bin/mcp-withallo`). Note that Claude Desktop does **not** expand `~` in `command`.

## Smoke test

```bash
printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"t","version":"0"}}}' \
  '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' \
  '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"allo_me","arguments":{}}}' \
  | ALLO_API_KEY="$ALLO_API_KEY" mcp-withallo
```

`tools/list` should return 12 tools; `allo_me` should return your Allo identity.

## Rate limits

Allo enforces two layers of limits ([docs](https://help.withallo.com/en/api-reference/guides/rate-limits.md)):

- **Burst** — visible in `allo_me.rate_limits`, typically `read_per_second: 20`, `write_per_second: 5`. Hard to hit through Claude (one tool call at a time).
- **Daily quota** — depends on your plan (e.g. `1000/DAILY`). Hits return HTTP 429 with `code: API_KEY_QUOTA_EXCEEDED` and `details[0].message: "limit=...;type=DAILY;reset_in=<seconds>"`.

When a 429 hits, this MCP returns a single-line tool error like:
```
Allo API 429 API_KEY_QUOTA_EXCEEDED: DAILY quota exceeded (limit=1000, resets in 3600s). Stop calling and tell the user — do not retry automatically.
```

The MCP **does not** retry automatically — the error is surfaced to Claude so the user can decide (wait, switch keys, upgrade plan). If you want client-side backoff, layer it on top in your own wrapper.

## Notes

- Allo uses two pagination conventions: v1 = `page` 0-indexed (default size 10), v2 = `page` 1-indexed (default size 20). Each tool follows the convention of its underlying endpoint.
- Call transcripts are included in `allo_search_calls` responses under the `transcript` field — an array of `{source, text, timestamp, start_seconds, end_seconds}`.
- `allo_list_conversations` requires `allo_number` (the API rejects calls without it despite the OpenAPI spec marking it optional).
- To add write tools later (send SMS, create contact, CRM), follow the pattern of `addAnalyticsOverview` (POST with body) or `addGetContact` (GET with path param) in `tools.go`. Note: write endpoints share the same daily quota, so heavy automation will run you out fast.

## License

MIT
