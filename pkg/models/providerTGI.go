package models

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"qatai/pkg/db"
	"strings"
	"time"
)

type TgiRequestBody struct {
	Inputs     string        `json:"inputs"`
	Parameters TgiParameters `json:"parameters"`
}

type TgiParameters struct {
	MaxNewTokens      int      `json:"max_new_tokens"`     //required
	Temperature       float64  `json:"temperature"`        //required
	TopK              int      `json:"top_k"`              //required
	TopP              float64  `json:"top_p"`              //required
	Stop              []string `json:"stop"`               //required
	RepetitionPenalty float64  `json:"repetition_penalty"` //required
	// BestOf              int      `json:"best_of"`
	// DecoderInputDetails bool    `json:"decoder_input_details"`
	Details bool `json:"details"`
	// DoSample            bool    `json:"do_sample"`
	ReturnFullText bool    `json:"return_full_text"`
	Seed           *int    `json:"seed"`
	Truncate       *int    `json:"truncate"`
	TypicalP       float64 `json:"typical_p"`
	Watermark      bool    `json:"watermark"`
}

// TGI response structs

type tgiResponse struct {
	Token struct {
		Id      int     `json:"id"`
		Text    string  `json:"text"`
		Logprob float64 `json:"logprob"`
		Special bool    `json:"special"`
	} `json:"token"`
	Generated_text string `json:"generated_text"`

	Details struct {
		FinishReason    string `json:"finish_reason"`
		GeneratedTokens int    `json:"generated_tokens"`
		// Seed            int64  `json:"seed"`
	} `json:"details"`
}

func replaceUniversalRoleWithModelRole(uni_role ChatRole, model *db.LLMModel) string {
	switch uni_role {
	case SYSTEM_ROLE:
		return model.Tokens.SystemToken
	case ASSISTANT_ROLE:
		return model.Tokens.AssistantToken
	case USER_ROLE:
		return model.Tokens.UserToken
	case FUNCTION_ROLE:
		return model.Tokens.FunctionToken
	}
	return ""
}
func convertUniversalRequestToTGI(req *UniversalRequest, model *db.LLMModel) *TgiRequestBody {
	//lets generate the INPUT message first
	inputText := ""

	for _, m := range req.Messages {
		inputText += replaceUniversalRoleWithModelRole(ChatRole(m.Role), model) + " "
		inputText += m.Content + " \n"
	}

	data := TgiRequestBody{
		Inputs: inputText,
		Parameters: TgiParameters{
			MaxNewTokens:      int(model.Parameters.MaxNewTokens),  //required
			RepetitionPenalty: model.Parameters.RepetitionPenality, //required
			Stop:              model.Stops,                         //required
			Temperature:       model.Parameters.Temperature,        //required
			TopK:              model.Parameters.Top_K,              //required
			TopP:              0.95,                                //required
			TypicalP:          0.95,                                //required
			// BestOf:              1,
			// DecoderInputDetails: false,
			Details: true,
			// DoSample: false,

			// ReturnFullText:    true,
			// Seed:              nil,

			// Truncate:          nil,

			// Watermark:         false,
		},
	}
	return &data
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func randSeq(n int) string {
	b := make([]rune, n)

	for i := range b {
		val, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		b[i] = letters[val.Int64()]
	}

	return string(b)
}
func DoGenerate(uReq *UniversalRequest, model *db.LLMModel, uRespChan chan string) *UniversalResponse {

	url := "http://gpu01.yawal.io:8080/generate"
	if uReq.Stream {
		url = "http://gpu01.yawal.io:8080/generate_stream"
	}
	//generate random hash
	random_uuid := randSeq(29)
	data := convertUniversalRequestToTGI(uReq, model)
	//TODO: check if we can add it if its missing, I mean if last role was not assitant, we add that
	data.Inputs += replaceUniversalRoleWithModelRole(ASSISTANT_ROLE, model) + " "
	jsonRequest, err := json.Marshal(data)

	if err != nil {
		log.Fatalln(err)
	}
	// log.Println(string(jsonRequest))
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonRequest))
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	//in openAI, we have to parse output in a different way, if we are doing it by stream or non stream
	// with TGI, this works either way, Thanks GPT-4
	if uReq.Stream {
		reader := bufio.NewReader(resp.Body)

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					//in case of stream, just to follow open ai standard, we have to send: "data: [DONE]"
					if uRespChan != nil {
						uRespChan <- "[DONE]"
					}
					break
				}
				log.Fatalln(err)
			}

			pos := strings.Index(line, "{")
			if pos != -1 {
				var tgiResp tgiResponse
				dataToParse := line[pos:]
				err := json.Unmarshal([]byte(dataToParse), &tgiResp)
				if err != nil {
					log.Fatalln(err)
				}
				// process data here
				uresp := convertTGIResponseToUniversalResponse(&tgiResp, random_uuid, model, uReq.Stream)
				uresp_data, err := json.Marshal(uresp)
				if err != nil {
					log.Fatalln(err)
				}
				// log.Println(string(uresp_data))
				if uRespChan != nil {
					uRespChan <- string(uresp_data)
				}
			}
		}
	} else {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}
		// log.Println(string(body))
		tgiResponseObj := new(tgiResponse)
		err = json.Unmarshal(body, &tgiResponseObj)
		if err != nil {
			log.Fatalln(err)
		}
		uresp := convertTGIResponseToUniversalResponse(tgiResponseObj, random_uuid, model, uReq.Stream)
		uresp_data, err := json.Marshal(uresp)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println(string(uresp_data))
	}

	return nil
}

func convertTGIResponseToUniversalResponse(resp *tgiResponse, randomUUID string, model *db.LLMModel, IsStream bool) *UniversalResponse {
	//filter out all unwanted characters from the string
	resp.Token.Text = strings.ReplaceAll(resp.Token.Text, "</s>", "") // TODO: check if this is ok later
	var finiedh_reason *string
	finiedh_reason = &resp.Details.FinishReason
	//strickly following open ai here
	if resp.Details.FinishReason == "" {
		finiedh_reason = nil
	}
	if resp.Details.FinishReason == "eos_token" {
		*finiedh_reason = "stop"
	}
	message := &Message{
		Role:    string(ASSISTANT_ROLE),
		Content: resp.Token.Text,
	}
	delta := &Delta{
		Content: &resp.Token.Text,
	}
	if IsStream {
		message = nil
	}
	if !IsStream {
		delta = nil
	}

	// if *finiedh_reason == "stop" {
	// 	delta.Content = nil
	// }
	ures := &UniversalResponse{
		ID:      fmt.Sprintf("chatcmpl-%s", randomUUID),
		Object:  "chat.completion.chunk",
		Created: int(time.Now().Unix()),
		Model:   model.Name,
		Usage:   nil, //not gonna do that
		Choices: []Choice{
			{
				Index:        0,
				Message:      message,
				Delta:        delta,
				FinishReason: finiedh_reason,
			},
		},
	}

	return ures
}

//TGI
/*

2023/07/26 06:19:54 {"token":{"id":4234,"text":" country","logprob":0.0,"special":false},"generated_text":null,"details":null}

2023/07/26 06:19:54 {"token":{"id":297,"text":" in","logprob":0.0,"special":false},"generated_text":null,"details":null}

2023/07/26 06:19:54 {"token":{"id":278,"text":" the","logprob":0.0,"special":false},"generated_text":"The biggest country in the","details":{"finish_reason":"length","generated_tokens":5,"seed":1858748514924990677}}
*/

// OpenAI
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
