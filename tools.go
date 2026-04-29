package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// registerTools wires every read-only Allo endpoint to a tool on s.
func registerTools(s *server.MCPServer, c *AlloClient) {
	addMe(s, c)
	addListNumbers(s, c)
	addListUsers(s, c)
	addListTags(s, c)
	addSearchCalls(s, c)
	addSearchContacts(s, c)
	addGetContact(s, c)
	addGetContactConversation(s, c)
	addListConversations(s, c)
	addSearchConversationItems(s, c)
	addGetConversationItem(s, c)
	addAnalyticsOverview(s, c)
}

// respond turns a raw HTTP response into a CallToolResult.
// 2xx → text content with the JSON body verbatim. Non-2xx → tool-level error.
// 429 is formatted explicitly so Claude can decide whether to back off.
func respond(body []byte, status int, err error) (*mcp.CallToolResult, error) {
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("HTTP error: %v", err)), nil
	}
	if status == 429 {
		return mcp.NewToolResultError(formatRateLimit(body)), nil
	}
	if status < 200 || status >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("Allo API %d: %s", status, string(body))), nil
	}
	if len(body) == 0 {
		return mcp.NewToolResultText("{}"), nil
	}
	return mcp.NewToolResultText(string(body)), nil
}

// formatRateLimit parses the 429 body shape documented at
// https://help.withallo.com/en/api-reference/guides/rate-limits and produces
// a single sentence Claude can act on. Falls back to the raw body on parse failure.
func formatRateLimit(body []byte) string {
	var parsed struct {
		Code    string `json:"code"`
		Error   struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
		Details []struct {
			Message string `json:"message"`
			Field   string `json:"field"`
		} `json:"details"`
	}
	_ = json.Unmarshal(body, &parsed)
	code := parsed.Code
	if code == "" {
		code = parsed.Error.Code
	}
	if code == "" {
		return fmt.Sprintf("Allo API 429 (rate limit): %s", string(body))
	}
	limit, kind, resetIn := "?", "?", "?"
	if len(parsed.Details) > 0 {
		for kv := range strings.SplitSeq(parsed.Details[0].Message, ";") {
			k, v, ok := strings.Cut(strings.TrimSpace(kv), "=")
			if !ok {
				continue
			}
			switch k {
			case "limit":
				limit = v
			case "type":
				kind = v
			case "reset_in":
				resetIn = v
			}
		}
	}
	return fmt.Sprintf("Allo API 429 %s: %s quota exceeded (limit=%s, resets in %ss). Stop calling and tell the user — do not retry automatically.", code, kind, limit, resetIn)
}

// --- v2 simple GETs ---------------------------------------------------------

func addMe(s *server.MCPServer, c *AlloClient) {
	tool := mcp.NewTool("allo_me",
		mcp.WithDescription("Get the authenticated Allo account identity. Use this as a sanity check that the API key works."),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return respond(c.Get(ctx, "/v2/api/me", nil))
	})
}

func addListNumbers(s *server.MCPServer, c *AlloClient) {
	tool := mcp.NewTool("allo_list_numbers",
		mcp.WithDescription("List the Allo phone numbers (E.164) attached to the account. Call this first when you need an `allo_number` parameter for `allo_search_calls` or `allo_list_conversations`."),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return respond(c.Get(ctx, "/v2/api/numbers", nil))
	})
}

func addListUsers(s *server.MCPServer, c *AlloClient) {
	tool := mcp.NewTool("allo_list_users",
		mcp.WithDescription("List team members on the Allo account."),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return respond(c.Get(ctx, "/v2/api/users", nil))
	})
}

func addListTags(s *server.MCPServer, c *AlloClient) {
	tool := mcp.NewTool("allo_list_tags",
		mcp.WithDescription("List all tags configured on the Allo account."),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return respond(c.Get(ctx, "/v2/api/tags", nil))
	})
}

// --- v1 calls + contacts ----------------------------------------------------

func addSearchCalls(s *server.MCPServer, c *AlloClient) {
	tool := mcp.NewTool("allo_search_calls",
		mcp.WithDescription("Search call history for an Allo number. Each result includes the full transcript (array of {source,text,timestamp,start_seconds,end_seconds}), AI summary, recording_url, tag, and direction (INBOUND/OUTBOUND). Page is 0-indexed."),
		mcp.WithString("allo_number",
			mcp.Required(),
			mcp.Description("Your Allo phone number in E.164 format (e.g., +14155551234). Use allo_list_numbers to discover it."),
		),
		mcp.WithString("contact_number",
			mcp.Description("Optional contact phone number (E.164) to narrow results to calls with that contact."),
		),
		mcp.WithNumber("page",
			mcp.DefaultNumber(0),
			mcp.Description("Page number, 0-indexed. Default 0."),
		),
		mcp.WithNumber("size",
			mcp.DefaultNumber(10),
			mcp.Min(1), mcp.Max(100),
			mcp.Description("Results per page (1-100). Default 10."),
		),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		alloNumber, err := req.RequireString("allo_number")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		q := url.Values{}
		q.Set("allo_number", alloNumber)
		if v := req.GetString("contact_number", ""); v != "" {
			q.Set("contact_number", v)
		}
		q.Set("page", strconv.Itoa(req.GetInt("page", 0)))
		q.Set("size", strconv.Itoa(req.GetInt("size", 10)))
		return respond(c.Get(ctx, "/v1/api/calls", q))
	})
}

func addSearchContacts(s *server.MCPServer, c *AlloClient) {
	tool := mcp.NewTool("allo_search_contacts",
		mcp.WithDescription("List contacts with engagement level. Page is 0-indexed."),
		mcp.WithNumber("page",
			mcp.DefaultNumber(0),
			mcp.Description("Page number, 0-indexed. Default 0."),
		),
		mcp.WithNumber("size",
			mcp.DefaultNumber(10),
			mcp.Min(1), mcp.Max(100),
			mcp.Description("Results per page (1-100). Default 10."),
		),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		q := url.Values{}
		q.Set("page", strconv.Itoa(req.GetInt("page", 0)))
		q.Set("size", strconv.Itoa(req.GetInt("size", 10)))
		return respond(c.Get(ctx, "/v1/api/contacts", q))
	})
}

func addGetContact(s *server.MCPServer, c *AlloClient) {
	tool := mcp.NewTool("allo_get_contact",
		mcp.WithDescription("Retrieve a single contact by ID (e.g., 'cnt_abc123'). Pass `extend` to include additional sections."),
		mcp.WithString("contact_id",
			mcp.Required(),
			mcp.Description("Contact ID, e.g. 'cnt_abc123'."),
		),
		mcp.WithString("extend",
			mcp.Description("Comma-separated extras: 'engagement', 'activity', or both."),
		),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("contact_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		q := url.Values{}
		if v := req.GetString("extend", ""); v != "" {
			q.Set("extend", v)
		}
		return respond(c.Get(ctx, "/v1/api/contact/"+url.PathEscape(id), q))
	})
}

func addGetContactConversation(s *server.MCPServer, c *AlloClient) {
	tool := mcp.NewTool("allo_get_contact_conversation",
		mcp.WithDescription("Get the calls + SMS exchanged with a specific contact, including recordings, transcripts, and AI summaries. Page is 0-indexed."),
		mcp.WithString("contact_id",
			mcp.Required(),
			mcp.Description("Contact ID, e.g. 'cnt_abc123'."),
		),
		mcp.WithString("allo_number",
			mcp.Description("Optional Allo phone number (E.164) to filter the conversation."),
		),
		mcp.WithNumber("page",
			mcp.DefaultNumber(0),
			mcp.Description("Page number, 0-indexed. Default 0."),
		),
		mcp.WithNumber("size",
			mcp.DefaultNumber(10),
			mcp.Min(1), mcp.Max(100),
			mcp.Description("Results per page (1-100). Default 10."),
		),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("contact_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		q := url.Values{}
		if v := req.GetString("allo_number", ""); v != "" {
			q.Set("allo_number", v)
		}
		q.Set("page", strconv.Itoa(req.GetInt("page", 0)))
		q.Set("size", strconv.Itoa(req.GetInt("size", 10)))
		return respond(c.Get(ctx, "/v1/api/contact/"+url.PathEscape(id)+"/conversation", q))
	})
}

// --- v2 conversations + analytics ------------------------------------------

func addListConversations(s *server.MCPServer, c *AlloClient) {
	tool := mcp.NewTool("allo_list_conversations",
		mcp.WithDescription("List conversations grouped by contact phone number, sorted by most recent activity. Page is 1-indexed (v2 convention)."),
		mcp.WithString("allo_number",
			mcp.Required(),
			mcp.Description("Allo phone number (E.164). Required: the API rejects this call without one."),
		),
		mcp.WithString("last_activity_since",
			mcp.Description("ISO 8601 timestamp; only return conversations active after this. Useful for incremental syncs."),
		),
		mcp.WithBoolean("unread",
			mcp.Description("If true, return only unread conversations."),
		),
		mcp.WithNumber("page",
			mcp.DefaultNumber(1),
			mcp.Description("Page number, 1-indexed. Default 1."),
		),
		mcp.WithNumber("size",
			mcp.DefaultNumber(20),
			mcp.Min(1), mcp.Max(100),
			mcp.Description("Results per page (1-100). Default 20."),
		),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		alloNumber, err := req.RequireString("allo_number")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		q := url.Values{}
		q.Set("allo_number", alloNumber)
		if v := req.GetString("last_activity_since", ""); v != "" {
			q.Set("last_activity_since", v)
		}
		if req.GetBool("unread", false) {
			q.Set("unread", "true")
		}
		q.Set("page", strconv.Itoa(req.GetInt("page", 1)))
		q.Set("size", strconv.Itoa(req.GetInt("size", 20)))
		return respond(c.Get(ctx, "/v2/api/conversations", q))
	})
}

func addSearchConversationItems(s *server.MCPServer, c *AlloClient) {
	tool := mcp.NewTool("allo_search_conversation_items",
		mcp.WithDescription("Keyword search across all calls and SMS. Terms are AND'd with prefix matching ('bill' matches 'billing'). Searches transcripts, summaries, and SMS content. Use this when the user asks 'find calls about X' or 'show messages mentioning Y'."),
		mcp.WithString("search",
			mcp.Description("Keyword query (not natural language). AND'd with prefix matching."),
		),
		mcp.WithString("allo_number", mcp.Description("Filter by Allo number (E.164).")),
		mcp.WithString("contact_number", mcp.Description("Filter by contact number (E.164).")),
		mcp.WithString("user_id", mcp.Description("Filter by user ID.")),
		mcp.WithString("direction",
			mcp.Description("INBOUND or OUTBOUND."),
			mcp.Enum("INBOUND", "OUTBOUND"),
		),
		mcp.WithString("type",
			mcp.Description("CALL, SMS, or ALL. Default ALL."),
			mcp.Enum("CALL", "SMS", "ALL"),
		),
		mcp.WithString("result",
			mcp.Description("ANSWERED, VOICEMAIL, or TRANSFERRED."),
			mcp.Enum("ANSWERED", "VOICEMAIL", "TRANSFERRED"),
		),
		mcp.WithBoolean("unread", mcp.Description("Only items unread by the team.")),
		mcp.WithBoolean("unresponded", mcp.Description("Only items still awaiting a response.")),
		mcp.WithString("sort",
			mcp.Description("DATE_DESC, DATE_ASC, or RELEVANCE. Default DATE_DESC."),
			mcp.Enum("DATE_DESC", "DATE_ASC", "RELEVANCE"),
		),
		mcp.WithString("date_from", mcp.Description("ISO 8601 start of date range filter.")),
		mcp.WithString("date_to", mcp.Description("ISO 8601 end of date range filter.")),
		mcp.WithNumber("page",
			mcp.DefaultNumber(1),
			mcp.Description("Page number, 1-indexed. Default 1."),
		),
		mcp.WithNumber("size",
			mcp.DefaultNumber(20),
			mcp.Min(1), mcp.Max(100),
			mcp.Description("Results per page (1-100). Default 20."),
		),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		body := map[string]any{
			"page": req.GetInt("page", 1),
			"size": req.GetInt("size", 20),
		}
		setIfStr := func(key string) {
			if v := req.GetString(key, ""); v != "" {
				body[key] = v
			}
		}
		for _, k := range []string{"search", "allo_number", "contact_number", "user_id", "direction", "type", "result", "sort"} {
			setIfStr(k)
		}
		if req.GetBool("unread", false) {
			body["unread"] = true
		}
		if req.GetBool("unresponded", false) {
			body["unresponded"] = true
		}
		df, dt := req.GetString("date_from", ""), req.GetString("date_to", "")
		if df != "" || dt != "" {
			date := map[string]string{}
			if df != "" {
				date["from"] = df
			}
			if dt != "" {
				date["to"] = dt
			}
			body["date"] = date
		}
		return respond(c.PostJSON(ctx, "/v2/api/conversations/items/search", body))
	})
}

func addGetConversationItem(s *server.MCPServer, c *AlloClient) {
	tool := mcp.NewTool("allo_get_conversation_item",
		mcp.WithDescription("Fetch a single call ('cll-*' ID) or SMS ('msg-*' ID) with full details (transcript, summary, recording_url for calls; body for SMS)."),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Item ID. 'cll-*' for calls, 'msg-*' for SMS."),
		),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return respond(c.Get(ctx, "/v2/api/conversations/items/"+url.PathEscape(id), nil))
	})
}

func addAnalyticsOverview(s *server.MCPServer, c *AlloClient) {
	tool := mcp.NewTool("allo_analytics_overview",
		mcp.WithDescription("Team analytics over a date range: call volume, SMS volume, average handling, etc."),
		mcp.WithString("date_from",
			mcp.Required(),
			mcp.Description("ISO 8601 start of date range, e.g. '2026-04-01'."),
		),
		mcp.WithString("date_to",
			mcp.Required(),
			mcp.Description("ISO 8601 end of date range, e.g. '2026-04-29'."),
		),
		mcp.WithString("allo_number",
			mcp.Description("Optional Allo number (E.164) to scope metrics."),
		),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		from, err := req.RequireString("date_from")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		to, err := req.RequireString("date_to")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		body := map[string]any{
			"date": map[string]string{"from": from, "to": to},
		}
		if v := req.GetString("allo_number", ""); v != "" {
			body["allo_number"] = v
		}
		return respond(c.PostJSON(ctx, "/v2/api/analytics/overview", body))
	})
}

