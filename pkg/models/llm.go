package models

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/bodaay/qatai/pkg/db"
)

// Initialise LLMModel
func NewLLMModel(name, description, prePrompt string, endpoints []db.LLMEndPoint, prompts []db.LLMPrompts, params db.LLMParameters) *db.LLMModel {
	return &db.LLMModel{
		Name:        name,
		Description: description,
		PrePrompt:   prePrompt,
		EndPoints:   endpoints,
		Prompts:     prompts,
		Parameters:  params,
	}
}

// ModelInfoResponse Seriously who is writing these anymore, just give the respones from curl request and ask chatGPT to write this
type TgiModelEndPoint struct {
	ModelID               string    `json:"model_id"`
	ModelSha              string    `json:"model_sha"`
	ModelDtype            string    `json:"model_dtype"`
	ModelDeviceType       string    `json:"model_device_type"`
	ModelPipelineTag      string    `json:"model_pipeline_tag"`
	MaxConcurrentRequests int       `json:"max_concurrent_requests"`
	MaxBestOf             int       `json:"max_best_of"`
	MaxStopSequences      int       `json:"max_stop_sequences"`
	MaxInputLength        int       `json:"max_input_length"`
	MaxTotalTokens        int       `json:"max_total_tokens"`
	WaitingServedRatio    float64   `json:"waiting_served_ratio"`
	MaxBatchTotalTokens   int       `json:"max_batch_total_tokens"`
	MaxWaitingTokens      int       `json:"max_waiting_tokens"`
	ValidationWorkers     int       `json:"validation_workers"`
	Version               string    `json:"version"`
	DBModelUUID           string    `json:"-"`
	LastChecked           time.Time `json:"-"`
	IsActive              bool      `json:"-"`
	Host                  string    `json:"-"`
	port                  int16     `json:"-"`
	UseSSL                bool      `json:"-"`
	DontSkipVerifySSL     bool      `json:"-"`

	// Sha                   string  `json:"sha"`
	// DockerLabel           string  `json:"docker_label"`
}

func GetTgiModel(uuid string, host string, port int16, useSSL bool, skipVerify bool) (*TgiModelEndPoint, error) {

	scheme := "http"
	if useSSL {
		scheme = "https"
	}

	requestUrl := &url.URL{
		Scheme: scheme,
		Host:   fmt.Sprintf("%s:%d", host, port),
	}

	// Skip TLS verification if necessary
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipVerify},
	}

	client := &http.Client{Transport: tr}

	response, err := client.Get(requestUrl.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get response: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var infoResponse TgiModelEndPoint
	if err := json.Unmarshal(body, &infoResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal json: %w", err)
	}
	infoResponse.DBModelUUID = uuid
	infoResponse.Host = host
	infoResponse.port = port
	infoResponse.DontSkipVerifySSL = skipVerify
	return &infoResponse, nil
}
