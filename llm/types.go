package llm

type FinalJSONResponse struct {
	Logs []struct {
		Type               string `json:"type"`
		Level              string `json:"level"`
		Time               string `json:"time"`
		Message            string `json:"message"`
		OriginalLog        string `json:"original_log,omitempty"`
		LLMInvestigation   string `json:"llm_investigation,omitempty"`
		LLMSuggestedAction string `json:"llm_suggested_action,omitempty"`
	} `json:"logs"`
	ContextForNextLLMPeriodicCheck string `json:"context_for_next_llm_periodic_check"`
}

type Message struct {
	Role             string `json:"role"`
	Content          string `json:"content,omitempty"`
	ReasoningContent string `json:"reasoning_content,omitempty"` // for chain-of-thought reasoning
	// For tool call messages
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

type LLMResponse struct {
	Choices []struct {
		FinishReason string  `json:"finish_reason"`
		Message      Message `json:"message"`
	} `json:"choices"`
}

type ToolCall struct {
	Function ToolFunction `json:"function"`
	Id       string       `json:"id"`
	Index    int          `json:"index"`
	Type     string       `json:"type"`
}

type ToolFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type Tool struct {
	Type     string     `json:"type"`
	Function ToolSchema `json:"function"`
}

type ToolSchema struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}
