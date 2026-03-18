package linear

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const apiURL = "https://api.linear.app/graphql"

type response struct {
	Data struct {
		Issue struct {
			BranchName string `json:"branchName"`
			Title      string `json:"title"`
		} `json:"issue"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// Ticket holds the branch name and title for a Linear issue.
type Ticket struct {
	Branch string
	Title  string
}

// FetchTicket queries the Linear API for issue details.
func FetchTicket(apiKey, ticketID string) (*Ticket, error) {
	query := fmt.Sprintf(`{ "query": "{ issue(id: \"%s\") { branchName title } }" }`, ticketID)

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(query))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	var result response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("API error: %s", result.Errors[0].Message)
	}
	if result.Data.Issue.BranchName == "" {
		return nil, fmt.Errorf("no branch found for ticket %s — check the ticket ID and your API key", ticketID)
	}

	return &Ticket{
		Branch: result.Data.Issue.BranchName,
		Title:  result.Data.Issue.Title,
	}, nil
}
