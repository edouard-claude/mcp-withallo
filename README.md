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
      "command": "/Users/edouard/go/bin/mcp-withallo",
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
      "command": "/Users/edouard/go/bin/mcp-withallo",
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
      "command": "/Users/edouard/go/bin/mcp-withallo",
      "env": { "ALLO_API_KEY": "your-api-key-here" }
    }
  }
}
```

Replace `/Users/edouard/go/bin/mcp-withallo` with the actual path on your machine if it differs (run `which mcp-withallo` after `go install`).

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

## Notes

- Allo uses two pagination conventions: v1 = `page` 0-indexed (default size 10), v2 = `page` 1-indexed (default size 20). Each tool follows the convention of its underlying endpoint.
- Call transcripts are included in `allo_search_calls` responses under the `transcript` field — an array of `{source, text, timestamp, start_seconds, end_seconds}`.
- To add write tools later (send SMS, create contact, CRM), follow the pattern of `addAnalyticsOverview` (POST with body) or `addGetContact` (GET with path param) in `tools.go`.

## License

MIT
