package models

type ChatRole string

const (
	SYSTEM_ROLE    ChatRole = "system"
	ASSISTANT_ROLE ChatRole = "assistant"
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
	Stop             string            `json:"stop"`
	PresencePenalty  float64           `json:"presence_penalty"`
	FrequencyPenalty float64           `json:"frequency_penalty"`
	LogitBias        map[string]string `json:"logit_bias"`
}

// ****************** Universal Generate/Completion Response ***************************
type Delta struct {
	Content *string `json:"content,omitempty"`
}
type Choice struct {
	Index        int      `json:"index"`
	Message      *Message `json:"message,omitempty"` //this actually whats returned in case stream: false
	Delta        *Delta   `json:"delta,omitempty"`   //this actually whats returned in case its stream: true
	FinishReason *string  `json:"finish_reason"`
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
	Usage   *Usage   `json:"usage,omitempty"`
}

/* stream: false
{
  "id": "chatcmpl-7gRfHylYcsDGV9NLN3HEn3nRxoYxV",
  "object": "chat.completion",
  "created": 1690350475,
  "model": "gpt-3.5-turbo-0613",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Yes, the biggest country in the world by land area is Russia."
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 29,
    "completion_tokens": 14,
    "total_tokens": 43
  }
}
*/
/* stream: true
data: {"id":"chatcmpl-7gRfDOBCCH81IVYamIoqNpJehg6eD","object":"chat.completion.chunk","created":1690350471,"model":"gpt-3.5-turbo-0613","choices":[{"index":0,"delta":{"content":"."},"finish_reason":null}]}

data: {"id":"chatcmpl-7gRfDOBCCH81IVYamIoqNpJehg6eD","object":"chat.completion.chunk","created":1690350471,"model":"gpt-3.5-turbo-0613","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}

data: [DONE]
*/
