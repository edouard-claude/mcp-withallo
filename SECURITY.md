# Security policy

## Reporting a vulnerability

Please report security issues privately by emailing **edouard@squirrel.fr** with the subject `[mcp-withallo security]`. Do not open a public GitHub issue for vulnerabilities.

I aim to acknowledge reports within 72 hours and to ship a fix or mitigation within 14 days for confirmed issues.

## Scope

- **In scope:** vulnerabilities in this MCP server (credential leakage, prompt-injection vectors that escalate privileges, command injection through tool arguments, path traversal, etc.).
- **Out of scope:** issues in the upstream Allo API itself — please report those to Allo. Issues in [`mark3labs/mcp-go`](https://github.com/mark3labs/mcp-go) should be reported to that project.

## Handling your API key

`ALLO_API_KEY` is read from the environment and sent verbatim in the `Authorization` header. **Never commit it to git.** Never paste it into an issue, PR, or log gist. The `.gitignore` excludes common env files (`.env`, `*.local`) but you are ultimately responsible for keeping the key out of your commits.

If you suspect a key was leaked, rotate it immediately in the Allo dashboard.
