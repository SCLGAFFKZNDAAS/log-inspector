package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log-inspector/loki"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var httpClient = &http.Client{}

func chat(messages []Message, tools []Tool) (LLMResponse, error) {
	body, _ := json.Marshal(map[string]any{
		"model":    os.Getenv("MODEL_NAME"),
		"messages": messages,
		"tools":    tools,
		"stream":   false,
		"response_format": map[string]any{
			"type": "json_object",
		},
	})

	LLM_URL := os.Getenv("LLM_URL")
	req, err := http.NewRequest("POST", LLM_URL, bytes.NewReader(body))
	if err != nil {
		return LLMResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("LLM_API_KEY")))
	resp, err := httpClient.Do(req)
	if err != nil {
		return LLMResponse{}, err
	}
	defer resp.Body.Close()

	var result LLMResponse
	json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}

func PeriodicCheck(lastPeriodicCheckResult FinalJSONResponse) (FinalJSONResponse, error) {
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

	services := strings.Split(os.Getenv("SERVICES_TO_SEARCH"), ",")
	_, filename, _, _ := runtime.Caller(0)          // get full path of current file
	dir := filepath.Dir(filename)                   // directory of the current file
	path := filepath.Join(dir, "system_prompt.txt") // file in the same directory
	sysPrompt, err := os.ReadFile(path)
	if err != nil {
		return FinalJSONResponse{}, fmt.Errorf("error reading system prompt: %v", err)
	}
	messages := []Message{
		{Role: "system", Content: string(sysPrompt)},
		{Role: "user", Content: fmt.Sprintf("TODAYS DATE IS %s. The available service_names are the following %v. Can you check if anything weird happened in the logs in the last 5 minutes? Here's the last periodic check result: %v", time.Now().Format(time.RFC3339), services, lastPeriodicCheckResult)},
	}

	assistantMsg, err := chat(messages, tools)
	if err != nil {
		return FinalJSONResponse{}, err
	}
	LLMmessage := assistantMsg.Choices[0].Message
	messages = append(messages, LLMmessage)

	for assistantMsg.Choices[0].FinishReason == "tool_calls" {
		for _, tc := range LLMmessage.ToolCalls {
			result := handleToolCall(tc)
			messages = append(messages, Message{
				Role:      "tool",
				Content:   result,
				ToolCalls: []ToolCall{tc},
			})
		}
		assistantMsg, err = chat(messages, tools)
		if err != nil {
			return FinalJSONResponse{}, err
		}
		LLMmessage = assistantMsg.Choices[0].Message
		messages = append(messages, LLMmessage)
	}

	var finalResponse FinalJSONResponse
	err = json.Unmarshal([]byte(assistantMsg.Choices[0].Message.Content), &finalResponse)
	if err != nil {
		return FinalJSONResponse{}, err
	}
	return finalResponse, nil
}

func handleToolCall(tc ToolCall) string {
	switch tc.Function.Name {
	case "query_logs":
		parsedArgs := make(map[string]any)
		err := json.Unmarshal([]byte(tc.Function.Arguments), &parsedArgs)
		if err != nil {
			return fmt.Sprintf("Error parsing arguments: %v", err)
		}
		query := parsedArgs["query"].(string)
		start := parsedArgs["start"].(string)

		var end string
		if parsedArgs["end"] == nil {
			end = time.Now().UTC().Format(time.RFC3339)
		} else {
			end = parsedArgs["end"].(string)
		}

		limit := int(parsedArgs["limit"].(float64))

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

		result, err := loki.QueryLoki(lokiQuery)
		if err != nil {
			return fmt.Sprintf("Error querying Loki: %v", err)
		}

		jsonResult, err := json.Marshal(result.Data.Result)
		if err != nil {
			return fmt.Sprintf("Error marshaling result: %v", err)
		}
		return string(jsonResult)
	default:
		return fmt.Sprintf("Unknown tool: %s", tc.Function.Name)
	}
}
