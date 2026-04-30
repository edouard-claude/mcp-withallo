<div align="center">

# mcp-withallo

**Bring your Allo conversations into Claude — every call, transcript, AI summary, message and KPI, read-only.**

[![Go Reference](https://pkg.go.dev/badge/github.com/edouard-claude/mcp-withallo.svg)](https://pkg.go.dev/github.com/edouard-claude/mcp-withallo)
[![Go Report Card](https://goreportcard.com/badge/github.com/edouard-claude/mcp-withallo)](https://goreportcard.com/report/github.com/edouard-claude/mcp-withallo)
[![CI](https://github.com/edouard-claude/mcp-withallo/actions/workflows/ci.yml/badge.svg)](https://github.com/edouard-claude/mcp-withallo/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/edouard-claude/mcp-withallo?include_prereleases&sort=semver)](https://github.com/edouard-claude/mcp-withallo/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)
[![MCP](https://img.shields.io/badge/MCP-compatible-8A2BE2)](https://modelcontextprotocol.io)

</div>

`mcp-withallo` is a local [Model Context Protocol](https://modelcontextprotocol.io) server that exposes the [Allo API](https://help.withallo.com/en/api-reference/introduction) as **read-only** tools to any MCP-compatible AI client — Claude Desktop, Claude Code, Cursor, Windsurf, VS Code, Zed, Cline, Continue, and others.

Once installed, ask Claude in plain English to summarize calls, dig through SMS threads, search transcripts for keywords, or pull team analytics. The server never sends messages, mutates contacts, or writes to your CRM — it only reads.

## Features

- **12 read-only tools** covering calls, transcripts, AI summaries, messages, contacts, conversations, tags and analytics.
- **Full call transcripts** — speaker-attributed turns with timestamps, recording URL and AI summary, exposed via `allo_search_calls` and `allo_get_conversation_item`.
- **Full-text conversation search** — `allo_search_conversation_items` matches across transcripts, summaries and message bodies, with filters for direction, date range, tags, recording presence, unread / unresponded state.
- **Period-over-period analytics** — `allo_analytics_overview` returns call/SMS volumes and handling-time KPIs, optionally compared to a baseline period and scoped to specific users.
- **Safe by design** — read-only by construction; nothing the AI can call writes to your Allo account or sends a message to a customer.
- **Single static binary** — built in Go, no runtime dependency, runs as a stdio MCP child process.
- **Aware of Allo rate limits** — surfaces HTTP 429 quota errors as a single actionable sentence so the model knows to stop instead of retrying blindly.

## Why this exists

[Allo](https://withallo.com) is an AI phone system for sales and support teams: every call is recorded, transcribed, AI-summarised and synced to a CRM. That data is exactly the kind of context an AI assistant should see — *"summarize my last 5 calls with Acme", "find all calls where the customer asked about pricing", "how did the team perform this week?"*. This MCP server makes that data queryable from any LLM that speaks MCP, without giving the model any way to send a message or change account state.

It's positioned in the same space as MCP servers for Aircall, OpenPhone, Dialpad, Twilio, Gong and Fireflies, but specifically for the Allo stack.

## Example prompts

Once configured, just ask Claude (or your MCP client of choice):

- *"Summarize my last 5 calls with Acme Corp."*
- *"Find every call where the customer mentioned 'refund' or 'cancel' this month."*
- *"How many calls did the team take last week, broken down by user?"*
- *"Pull yesterday's call with +14155551234 and give me the transcript and action items."*
- *"Which contacts have unresponded messages older than 48 hours?"*
- *"What were the top objections raised in calls tagged 'demo' in April?"*
- *"List my unread voicemails and summarize each one."*
- *"Compare call volume this week vs. last week on +33123456789."*
- *"Show me the AI summary of every inbound call yesterday that went to voicemail."*

## Install

### Option A — `go install` (recommended)

```bash
go install github.com/edouard-claude/mcp-withallo@latest
```

The binary lands at `$GOBIN/mcp-withallo` (typically `~/go/bin/mcp-withallo`). Make sure that directory is on your `PATH` if you want to invoke it without an absolute path.

### Option B — prebuilt binary

Grab the archive matching your OS / architecture from the [Releases page](https://github.com/edouard-claude/mcp-withallo/releases) and put `mcp-withallo` somewhere on your `PATH`.

### Option C — build from source

```bash
git clone https://github.com/edouard-claude/mcp-withallo
cd mcp-withallo
make build      # or: go build -o mcp-withallo .
```

## Configuration

Get an API key from the Allo dashboard (**Settings → API keys**) and put it in the `ALLO_API_KEY` environment variable. The key is sent verbatim in the `Authorization` header — no `Bearer` prefix, that's intentional ([per Allo docs](https://help.withallo.com/en/api-reference/guides/authentication)).

The same JSON config block works for every MCP client; only the file path differs.

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

<details>
<summary><strong>Claude Desktop</strong></summary>

Edit `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or `%APPDATA%\Claude\claude_desktop_config.json` (Windows), add the block above, then fully quit Claude (⌘Q) and relaunch.

</details>

<details>
<summary><strong>Claude Code</strong></summary>

- **Global** — add the block to `~/.claude/settings.json`.
- **Per project** — create `.mcp.json` at the repo root with the same block.

</details>

<details>
<summary><strong>Cursor</strong></summary>

Edit `~/.cursor/mcp.json` (global) or `.cursor/mcp.json` at the project root, then reload the window.

</details>

<details>
<summary><strong>Windsurf</strong></summary>

Edit `~/.codeium/windsurf/mcp_config.json`, then reload Cascade.

</details>

<details>
<summary><strong>VS Code (with MCP support)</strong></summary>

```bash
code --add-mcp '{"name":"allo","command":"mcp-withallo","env":{"ALLO_API_KEY":"your-api-key-here"}}'
```

</details>

## Tools

All 12 tools are read-only — no SMS sending, no contact mutations, no CRM writes.

### Account & directory

| Tool | Description | Allo endpoint |
|---|---|---|
| `allo_me` | Authenticated identity, scopes, rate-limit budget | `GET /v2/api/me` |
| `allo_list_numbers` | Phone numbers / sender IDs on the account | `GET /v2/api/numbers` |
| `allo_list_users` | Team members (filter by `role`, `status`) | `GET /v2/api/users` |
| `allo_list_tags` | Configured call/conversation tags | `GET /v2/api/tags` |

### Calls (with transcripts)

| Tool | Description | Allo endpoint |
|---|---|---|
| `allo_search_calls` | Call history with full transcript, AI summary, recording URL, tags, type, result | `GET /v1/api/calls` |

### Conversations & messages

| Tool | Description | Allo endpoint |
|---|---|---|
| `allo_list_conversations` | Conversations grouped by contact, sorted by recency | `GET /v2/api/conversations` |
| `allo_search_conversation_items` | Full-text search across calls and messages with rich filters | `POST /v2/api/conversations/items/search` |
| `allo_get_conversation_item` | One call (`cll-*`) or message (`msg-*`) with full detail | `GET /v2/api/conversations/items/{id}` |

### Contacts

| Tool | Description | Allo endpoint |
|---|---|---|
| `allo_search_contacts` | List contacts with engagement signal | `GET /v1/api/contacts` |
| `allo_get_contact` | One contact, optionally with `engagement` and `activity` extensions | `GET /v1/api/contact/{id}` |
| `allo_get_contact_conversation` | All calls + messages exchanged with one contact | `GET /v1/api/contact/{id}/conversation` |

### Analytics

| Tool | Description | Allo endpoint |
|---|---|---|
| `allo_analytics_overview` | Team KPIs over a date range, with optional period-over-period comparison and per-user filtering | `POST /v2/api/analytics/overview` |

## Smoke test

```bash
printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"smoke","version":"0"}}}' \
  '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' \
  '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"allo_me","arguments":{}}}' \
  | ALLO_API_KEY="$ALLO_API_KEY" mcp-withallo
```

`tools/list` should return 12 tools and `allo_me` should return your Allo identity. The `make smoke-test` target runs the same recipe.

## Rate limits

Allo enforces two layers ([docs](https://help.withallo.com/en/api-reference/guides/rate-limits)):

- **Burst** — visible in `allo_me.rate_limits`, typically `read_per_second: 20`, `write_per_second: 5`. Hard to hit through Claude (one tool call at a time).
- **Daily quota** — depends on your plan (e.g. `1000/DAILY`). Hits return HTTP 429 with `code: API_KEY_QUOTA_EXCEEDED`.

When a 429 hits, this server surfaces a single-line tool error like:

```
Allo API 429 API_KEY_QUOTA_EXCEEDED: DAILY quota exceeded (limit=1000, resets in 3600s). Stop calling and tell the user — do not retry automatically.
```

The server **does not** retry automatically — the error reaches the model so the user can decide (wait, switch keys, upgrade). If you want client-side back-off, layer it on top.

## Notes & known divergences from the OpenAPI spec

- **Pagination conventions differ between API versions.** v1 endpoints (calls, contacts, contact-conversation) use `page` 0-indexed with default size 10. v2 endpoints (conversations list, items search) use `page` 1-indexed with default size 20. Each tool follows its underlying endpoint's convention.
- **Call transcripts** ride inside `allo_search_calls` and `allo_get_conversation_item` responses under the `transcript` field — an array of `{source, text, time, start_seconds, end_seconds}`.
- **`allo_list_conversations` requires `allo_number`** — the live API rejects calls without it despite the OpenAPI spec marking the field optional. Empirically verified, see commit `e3c1e9c`.
- **`allo_search_conversation_items` enums** are aligned to the live API, not the OpenAPI examples: `type ∈ {CALL, TEXT_MESSAGE}`, `result ∈ {ANSWERED, MISSED, TRANSFERRED, VOICEMAIL}`, `sort ∈ {RELEVANCE, RECENT, OLDEST}`.

## Development

```bash
make build         # build ./mcp-withallo
make test          # go test -race ./...
make lint          # gofmt + go vet (CI-equivalent)
make smoke-test    # end-to-end smoke against the real API (needs ALLO_API_KEY)
```

The codebase is intentionally tiny: `main.go` (entry point), `client.go` (HTTP client), `tools.go` (every MCP tool definition + handler). To add a new read-only tool, mirror an existing pattern — see [`CONTRIBUTING.md`](./CONTRIBUTING.md) for the recipe.

## Roadmap

Potential additions, all read-only and tracked in [issues](https://github.com/edouard-claude/mcp-withallo/issues):

- `allo_get_user` — single team member by ID.
- `allo_batch_get_conversation_items` — fetch up to 100 items by ID in one call.
- `allo_analytics_outbound` — outbound funnel & heatmap.
- CRM read tools — `crm/people`, `crm/companies`, `crm/deals` search & get.
- `allo_list_webhooks` — read-only listing of configured webhooks.

Want one of these (or something else)? [Open an issue](https://github.com/edouard-claude/mcp-withallo/issues/new/choose).

## Contributing

PRs welcome — see [CONTRIBUTING.md](./CONTRIBUTING.md) for build steps, the tool-addition pattern, and code style. For security issues, see [SECURITY.md](./SECURITY.md) — please **do not** open a public issue for vulnerabilities.

## Acknowledgements

Built on [`mark3labs/mcp-go`](https://github.com/mark3labs/mcp-go), the Go SDK for the Model Context Protocol. Inspired by the read-only design choices of [`getsentry/sentry-mcp`](https://github.com/getsentry/sentry-mcp) and [`github/github-mcp-server`](https://github.com/github/github-mcp-server).

## License

[MIT](./LICENSE) — © 2026 Edouard Claude.
