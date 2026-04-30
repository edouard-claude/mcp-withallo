# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed
- `allo_search_conversation_items` enums now match the live API: `type` accepts `CALL` / `TEXT_MESSAGE` (was `CALL` / `SMS` / `ALL`); `result` adds `MISSED`; `sort` is `RELEVANCE` / `RECENT` / `OLDEST` (was `DATE_DESC` / `DATE_ASC` / `RELEVANCE`). Previous values were silently rejected by Allo.
- `allo_search_calls` description: transcript entries carry `time`, not `timestamp`; the call direction field is `type` (not `direction`); added `MISSED`, `TRANSFERRED_AI`, `BLOCKED` to the documented `result` outcomes.
- `allo_analytics_overview` no longer sends an unsupported `allo_number` field in the body (was silently ignored by Allo).

### Added
- `allo_search_conversation_items` now exposes `has_recording` (bool) and `tags` (array of tag IDs) filters.
- `allo_analytics_overview` exposes `compare_date_from` / `compare_date_to` (period-over-period) and `user_ids` (filter by team member).
- `allo_list_users` exposes `role` and `status` filters.
- `allo_list_conversations` exposes `extend` to enrich responses with `engagement` / `activity`.
- `allo_get_conversation_item` exposes `extend` for transcript / tag enrichment.
- LICENSE (MIT), CONTRIBUTING, SECURITY, CHANGELOG, Makefile, `.goreleaser.yaml`, GitHub Actions CI + release workflows, issue / PR templates.

## [0.1.0] - 2026-04-29

Initial release.

### Added
- Local stdio MCP server wrapping the [Allo API](https://help.withallo.com/en/api-reference/introduction).
- 12 read-only tools:
  - `allo_me`, `allo_list_numbers`, `allo_list_users`, `allo_list_tags`
  - `allo_search_calls` (includes transcripts), `allo_search_contacts`
  - `allo_get_contact`, `allo_get_contact_conversation`
  - `allo_list_conversations`, `allo_search_conversation_items`, `allo_get_conversation_item`
  - `allo_analytics_overview`
- Explicit handling of HTTP 429 rate-limit responses with structured, single-line error messages.
- README documenting Claude Desktop and Claude Code configuration, smoke-test recipe, and Allo rate-limit semantics.

### Notes
- `allo_list_conversations` requires `allo_number` (the Allo API rejects calls without it despite the OpenAPI spec marking it optional).
- The MCP does not retry on 429 — the error is surfaced to the client so the user can decide.

[Unreleased]: https://github.com/edouard-claude/mcp-withallo/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/edouard-claude/mcp-withallo/releases/tag/v0.1.0
