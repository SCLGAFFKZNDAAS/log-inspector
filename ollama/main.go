package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log-inspector/loki"
	"net/http"
	"os"
	"time"
)

func chat(messages []Message, tools []Tool) (Message, error) {
	body, _ := json.Marshal(map[string]any{
		"model":    "qwen3.5:9b",
		"messages": messages,
		"tools":    tools,
		"stream":   false,
	})

	OLLAMA_URL := os.Getenv("OLLAMA_URL")
	resp, err := http.Post(OLLAMA_URL+"/api/chat", "application/json", bytes.NewReader(body))
	if err != nil {
		return Message{}, err
	}
	defer resp.Body.Close()

	var result struct {
		Message Message `json:"message"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return result.Message, nil
}

func Test() {
	tools := []Tool{
		{
			Type: "function",
			Function: ToolSchema{
				Name:        "query_logs",
				Description: "Query logs from Loki using LogQL. Returns log entries matching the query and time range.",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"query": map[string]any{
							"type":        "string",
							"description": "REQUIRED! The Loki query string, e.g. {app=\"nextcloud\"} |= \"\"",
						},
						"start": map[string]any{
							"type":        "string",
							"description": "REQUIRED! Start time in RFC3339 format, e.g. `2024-01-01T00:00:00Z`",
						},
						"end": map[string]any{
							"type":        "string",
							"description": "REQUIRED! End time in RFC3339 format, e.g. `2024-01-01T01:00:00Z`",
						},
						"limit": map[string]any{
							"type":        "integer",
							"description": "REQUIRED! Maximum number of log entries to return",
						},
					},
					"required": []string{"query", "start", "end", "limit"},
				},
			},
		},
	}

	messages := []Message{
		{Role: "user", Content: fmt.Sprintf("TODAYS DATE IS %s. Can you check if anything weird happened in the logs in the last 1 hour?", time.Now().Format(time.RFC3339))},
	}

	assistantMsg, err := chat(messages, tools)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	messages = append(messages, assistantMsg)

	if len(assistantMsg.ToolCalls) > 0 {
		for _, tc := range assistantMsg.ToolCalls {
			result := handleToolCall(tc)

			println("Tool call result:", result)

			messages = append(messages, Message{
				Role:     "tool",
				Content:  result,
				ToolName: tc.Function.Name,
			})
		}

		finalMsg, err := chat(messages, tools)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println(finalMsg.Content)
	}
}

func handleToolCall(tc ToolCall) string {
	switch tc.Function.Name {
	case "query_logs":
		query := "{app=\"nextcloud\"} |= \"\"" // tc.Function.Arguments["query"].(string)
		start := tc.Function.Arguments["start"].(string)

		var end string
		if tc.Function.Arguments["end"] == nil {
			end = time.Now().UTC().Format(time.RFC3339)
		} else {
			end = tc.Function.Arguments["end"].(string)
		}

		limit := int(tc.Function.Arguments["limit"].(float64))

		startDate, err := time.Parse(time.RFC3339, start)
		if err != nil {
			return fmt.Sprintf("Error parsing start date: %v", err)
		}
		endDate, err := time.Parse(time.RFC3339, end)
		if err != nil {
			return fmt.Sprintf("Error parsing end date: %v", err)
		}

		lokiQuery := loki.LokiQuery{
			Query: query,
			Start: startDate,
			End:   endDate,
			Limit: limit,
		}

		fmt.Print(lokiQuery)

		result, err := loki.QueryLoki(lokiQuery)

		jsonResult, err := json.Marshal(result.Data.Result)
		if err != nil {
			return fmt.Sprintf("Error marshaling result: %v", err)
		}
		return string(jsonResult)
	default:
		return fmt.Sprintf("Unknown tool: %s", tc.Function.Name)
	}
}
