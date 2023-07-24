package db

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

const collection_bucket_name_models = "models"

type LLMProvider string

const (
	OPENAI LLMProvider = "openai"
	HFTGI  LLMProvider = "tgi"
)

type LLMModel struct {
	UUID        string        `json:"uuid"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	PrePrompt   string        `json:"prePrompt"`
	Provider    LLMProvider   `json:"provider"`
	Prompts     []LLMPrompts  `json:"prompts"`
	Parameters  LLMParameters `json:"parameters"`
	EndPoints   []LLMEndPoint `json:"endPoints"`
}

type LLMPrompts struct {
	Title       string `json:"title"`
	Prompt      string `json:"prompt"`
	PromptImage string `json:"promptImage"`
}

type LLMParameters struct {
	Temperature        float64 `json:"temperature"`
	Top_P              float64 `json:"topP"`
	RepetitionPenality float64 `json:"repetitionPenality"`
	Truncate           int64   `json:"truncate"`
	MaxNewTokens       int64   `json:"maxNewTokens"`
}

type LLMEndPoint struct {
	Host   string `json:"host"`
	Port   int16  `json:"port"`
	UseSSL bool   `json:"useSSL"`
}

func AddUpdateModel(db QataiDatabase, m *LLMModel, updateIfExists bool) error {
	//lets check if it exists before
	records, err := db.GetAllRecordForCollectionBucket(collection_bucket_name_models)
	if err != nil {
		return err
	}

	//in case of models, the only important thing that we know we are not hitting same end point
	found := false
	host := ""
	port := int16(0)
	for _, r := range records {
		var rval LLMModel
		err := json.Unmarshal([]byte(r.Value), &rval)
		if err != nil {
			return err
		}

		for _, e := range rval.EndPoints {
			for _, ee := range m.EndPoints {
				if e.Host == ee.Host && e.Port == ee.Port {
					found = true
					host = e.Host
					port = e.Port
					break // this is useless, but anyway, its ok
				}
			}
		}
	}
	if found && !updateIfExists {
		return fmt.Errorf("another model has same end point, host: %s:%d ", host, port)
	}

	if m.UUID == "" {
		m.UUID = uuid.New().String()
	}

	marshaleld, err := json.Marshal(m)
	if err != nil {
		return err
	}

	err = db.SetValueByKeyName(collection_bucket_name_models, &QataiDatabaseRecord{Key: m.UUID, Value: string(marshaleld)})
	if err != nil {
		return err
	}
	return nil
}
func GetAllModels(db QataiDatabase) ([]LLMModel, error) {
	var mmodels []LLMModel
	values, err := db.GetAllRecordForCollectionBucket(collection_bucket_name_models)
	if err != nil {
		return nil, err
	}

	for _, v := range values {
		var m LLMModel
		err := json.Unmarshal([]byte(v.Value), &m)
		if err != nil {
			return nil, err
		}
		mmodels = append(mmodels, m)
	}

	return mmodels, nil
}

func GetModelByName(db QataiDatabase, modelName string) (*LLMModel, error) {
	models, err := GetAllModels(db)
	if err != nil {
		return nil, err
	}
	for _, m := range models {
		if m.Name == modelName {
			return &m, nil
		}
	}
	return nil, fmt.Errorf("no model found matching this name: %s", modelName)
}
func GetModelByUUID(db QataiDatabase, uuid string) (*LLMModel, error) {
	models, err := GetAllModels(db)
	if err != nil {
		return nil, err
	}
	for _, m := range models {
		if m.UUID == uuid {
			return &m, nil
		}
	}
	return nil, fmt.Errorf("no model found matching this uuid: %s", uuid)
}
func ClearAllModels(db QataiDatabase) error {
	return db.ClearAllRecordsInCollection(collection_bucket_name_models)
}
