package models

type ChatRole string

const (
	SYSTEM_ROLE    ChatRole = "system"
	ASSISTANT_TOLE ChatRole = "assistant"
	USER_ROLE      ChatRole = "user"
	FUNCTION_ROLE  ChatRole = "function"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ****************** Universal Generate/Completion Request ***************************

type UniversalRequest struct {
	Messages         []Message         `json:"messages"`
	Stream           bool              `json:"stream"`
	Model            string            `json:"model"`
	Temperature      float64           `json:"temperature"`
	TopP             float64           `json:"top_p"`
	N                int               `json:"n"`
	Stop             []string          `json:"stop"`
	PresencePenalty  float64           `json:"presence_penalty"`
	FrequencyPenalty float64           `json:"frequency_penalty"`
	LogitBias        map[string]string `json:"logit_bias"`
}

// ****************** Universal Generate/Completion Response ***************************

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type UniversalResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int      `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}
