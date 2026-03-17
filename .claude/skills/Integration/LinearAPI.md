# Linear GraphQL API Integration

## Overview

The `lw` script calls the Linear GraphQL API to resolve a ticket ID into a branch name and title. This is the only external API call in the codebase. It is only used in ticket mode — scratch and review modes skip the API entirely.

## Authentication

- Header: `Authorization: $LINEAR_API_KEY`
- The key is a bare API key, **not** a Bearer token
- Source: `LINEAR_API_KEY` environment variable (required for ticket mode, not needed otherwise)

## Endpoint

`POST https://api.linear.app/graphql`

## Query

A single GraphQL query fetching two fields:

```graphql
{
  issue(id: "TICKET-ID") {
    branchName
    title
  }
}
```

The query is constructed as a JSON payload via heredoc and sent with `curl -s`.

## Response handling

1. **API errors:** checks `.errors[0].message` via `jq`. If present, dies with the error message.
2. **Branch extraction:** `.data.issue.branchName` — if empty, the ticket ID is invalid or has no branch configured.
3. **Title extraction:** `.data.issue.title` — used for informational output and tmux window naming.

## Dependencies

Ticket mode requires `curl` and `jq` (checked via `require curl jq`). Scratch and review modes do not need these.

## Error modes

| Condition | Behavior |
|-----------|----------|
| `LINEAR_API_KEY` not set | `die` before API call |
| API returns error object | `die` with the error message |
| `branchName` is empty/null | `die` suggesting user check ticket ID and API key |
| Network failure | `curl` returns non-zero, caught by `set -e` |

## Design notes

- The script constructs the GraphQL query by interpolating the ticket ID directly into the JSON string. This is safe because ticket IDs are pre-validated as `^[A-Za-z]+-[0-9]+` and uppercased.
- Only `branchName` and `title` are fetched — no other Linear data is used.
- If extending the query (e.g. to fetch labels, assignees, status), the `jq` extraction and error handling must be updated to match.
