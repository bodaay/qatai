package models

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"qatai/pkg/db"
)

type TgiRequestBody struct {
	Inputs     string        `json:"inputs"`
	Parameters TgiParameters `json:"parameters"`
}

type TgiParameters struct {
	MaxNewTokens        int      `json:"max_new_tokens"`     //required
	Temperature         float64  `json:"temperature"`        //required
	TopK                int      `json:"top_k"`              //required
	TopP                float64  `json:"top_p"`              //required
	Stop                []string `json:"stop"`               //required
	RepetitionPenalty   float64  `json:"repetition_penalty"` //required
	BestOf              int      `json:"best_of"`
	DecoderInputDetails bool     `json:"decoder_input_details"`
	Details             bool     `json:"details"`
	DoSample            bool     `json:"do_sample"`
	ReturnFullText      bool     `json:"return_full_text"`
	Seed                *int     `json:"seed"`
	Truncate            *string  `json:"truncate"`
	TypicalP            float64  `json:"typical_p"`
	Watermark           bool     `json:"watermark"`
}

// TGI response structs
type tgiToken struct {
	ID      int     `json:"id"`
	Text    string  `json:"text"`
	Logprob float64 `json:"logprob"`
	Special bool    `json:"special"`
}

type tgiDetails struct {
	FinishReason    string `json:"finish_reason"`
	GeneratedTokens int    `json:"generated_tokens"`
	// Seed            int64      `json:"seed"`
	Prefill []int      `json:"prefill"`
	Tokens  []tgiToken `json:"tokens"`
}

type tgiResponse struct {
	GeneratedText string     `json:"generated_text"`
	Details       tgiDetails `json:"details"`
}

func replaceUniversalRoleWithModelRole(uni_role ChatRole, model *db.LLMModel) string {
	switch uni_role {
	case SYSTEM_ROLE:
		return model.Tokens.SystemToken
	case ASSISTANT_TOLE:
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
			MaxNewTokens:      int(model.Parameters.MaxNewTokens), //required
			RepetitionPenalty: req.FrequencyPenalty,               //required
			Stop:              model.Stops,                        //required
			Temperature:       model.Parameters.Temperature,       //required
			TopK:              model.Parameters.Top_K,             //required
			TopP:              0.95,                               //required
			TypicalP:          0.95,                               //required
			// BestOf:              1,
			// DecoderInputDetails: false,
			Details: true,
			// DoSample:            false,

			// ReturnFullText:    true,
			// Seed:              nil,

			// Truncate:          nil,

			// Watermark:         false,
		},
	}
	return &data
}
func DoGenerate(uReq *UniversalRequest, model *db.LLMModel) *UniversalResponse {
	url := "http://gpu01.yawal.io:8080/generate"
	data := convertUniversalRequestToTGI(uReq, model)
	//append one more of role assistant to prepare it for response
	data.Inputs += replaceUniversalRoleWithModelRole(ASSISTANT_TOLE, model) + " "
	jsonRequest, err := json.Marshal(data)

	if err != nil {
		log.Fatalln(err)
	}
	log.Println(string(jsonRequest))
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonRequest))
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	tgiResponseObj := new(tgiResponse)
	err = json.Unmarshal(body, &tgiResponseObj)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(tgiResponseObj)
	return nil
}
