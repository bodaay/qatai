package pdb

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
)

func InitPDB(app *pocketbase.PocketBase) error {
	//dont change the order of these please
	app.OnBeforeServe().Add(
		func(e *core.ServeEvent) error {
			if _, err := app.Dao().FindCollectionByNameOrId("prompts_template"); err != nil {
				err := initCollection_PromptTemplates(e.App)
				if err != nil {
					return err
				}
			}
			prompts_collection, err := app.Dao().FindCollectionByNameOrId("prompts_template")
			if err != nil {
				return err
			}
			if _, err := app.Dao().FindCollectionByNameOrId("providers"); err != nil {
				err = initCollection_Providers(app)
				if err != nil {
					return err
				}

			}
			providers_collection, err := app.Dao().FindCollectionByNameOrId("providers")
			if err != nil {
				return err
			}
			//Insert default recrods for providers
			// record := models.NewRecord(providers_collection)
			// record.Set("title", "Lorem ipsum")
			// record.Set("active", true)
			// record.Set("someOtherField", 123)
			// if err := app.Dao().SaveRecord(record); err != nil {
			// 	return err
			// }
			if _, err := app.Dao().FindCollectionByNameOrId("models"); err != nil {
				err = initCollection_Models(app, prompts_collection, providers_collection)
				if err != nil {
					return err
				}
			}
			models_collection, err := app.Dao().FindCollectionByNameOrId("models")
			if err != nil {
				return err
			}

			if _, err := app.Dao().FindCollectionByNameOrId("endpoints"); err != nil {
				err = initCollection_Endpoints(app, models_collection)
				if err != nil {
					return err
				}
			}
			_, err = app.Dao().FindCollectionByNameOrId("endpoints")
			if err != nil {
				return err
			}
			if _, err := app.Dao().FindCollectionByNameOrId("config"); err != nil {
				err = initCollection_Config(app)
				if err != nil {
					return err
				}
			}
			return nil
		})

	return nil
}

func InsertDefaultRecords(app core.App) error {

	// config
	// models_collection, err := app.Dao().FindCollectionByNameOrId("models")
	// if err != nil {
	// 	return err
	// }
	// config_collection, err := app.Dao().FindCollectionByNameOrId("config")
	// if err != nil {
	// 	return err
	// }

	return nil
}

/*
	type LLMModel struct {
		UUID        string        `json:"uuid"`
		Name        string        `json:"name"`
		Description string        `json:"description"`
		PrePrompt   string        `json:"prePrompt"`
		Tokens      LLMTokens     `json:"tokens"`
		Stops       []string      `json:"stops"`
		Provider    LLMProvider   `json:"provider"`
		Prompts     []LLMPrompts  `json:"prompts"`
		Parameters  LLMParameters `json:"parameters"`
		EndPoints   []LLMEndPoint `json:"endPoints"`
	}

	type LLMTokens struct {
		SystemToken    string `json:"systemToken"`
		UserToken      string `json:"userToken"`
		AssistantToken string `json:"assistantToken"`
		FunctionToken  string `json:"functionToken"`
	}

	type LLMPrompts struct {
		Title       string `json:"title"`
		Prompt      string `json:"prompt"`
		PromptImage string `json:"promptImage"`
	}

	type LLMParameters struct {
		Temperature        float64 `json:"temperature"`
		Top_P              float64 `json:"topP"`
		Top_K              int     `json:"topK"`
		RepetitionPenality float64 `json:"repetitionPenality"`
		Truncate           int64   `json:"truncate"`
		MaxNewTokens       int     `json:"maxNewTokens"`
	}
*/
func initCollection_PromptTemplates(app core.App) error {

	collection := &models.Collection{
		Indexes: types.JsonArray[string]{
			"CREATE UNIQUE INDEX `idx_pt_name` ON `prompts_template` (`name`)",
		},
	}

	form := forms.NewCollectionUpsert(app, collection)
	form.Name = "prompts_template"
	form.Type = models.CollectionTypeBase
	form.ListRule = nil
	form.ViewRule = types.Pointer("@request.auth.id != ''")
	form.CreateRule = types.Pointer("")
	form.UpdateRule = types.Pointer("@request.auth.id != ''")
	form.DeleteRule = nil
	form.Schema.AddField(&schema.SchemaField{
		Name:     "name",
		Type:     schema.FieldTypeText,
		Required: true,
		Unique:   true,
		Options: &schema.TextOptions{
			Max: types.Pointer(100),
		},
	})

	form.Schema.AddField(&schema.SchemaField{
		Name:     "system_token",
		Type:     schema.FieldTypeText,
		Required: false,
		Options: &schema.TextOptions{
			Max: types.Pointer(250),
		},
	})
	form.Schema.AddField(&schema.SchemaField{
		Name:     "user_token",
		Type:     schema.FieldTypeText,
		Required: false,
		Options: &schema.TextOptions{
			Max: types.Pointer(250),
		},
	})
	form.Schema.AddField(&schema.SchemaField{
		Name:     "assistant_token",
		Type:     schema.FieldTypeText,
		Required: false,
		Options: &schema.TextOptions{
			Max: types.Pointer(250),
		},
	})
	form.Schema.AddField(&schema.SchemaField{
		Name:     "function_token",
		Type:     schema.FieldTypeText,
		Required: false,
		Options: &schema.TextOptions{
			Max: types.Pointer(250),
		},
	})

	if err := form.Submit(); err != nil {
		return err
	}

	return nil

}
func initCollection_Endpoints(app core.App, models_collection *models.Collection) error {

	collection := &models.Collection{
		Indexes: types.JsonArray[string]{
			"CREATE UNIQUE INDEX `idx_ep_host` ON `endpoints` (`host`)",
		},
	}

	form := forms.NewCollectionUpsert(app, collection)
	form.Name = "endpoints"
	form.Type = models.CollectionTypeBase
	form.ListRule = nil
	form.ViewRule = types.Pointer("@request.auth.id != ''")
	form.CreateRule = types.Pointer("")
	form.UpdateRule = types.Pointer("@request.auth.id != ''")
	form.DeleteRule = nil
	form.Schema.AddField(&schema.SchemaField{
		Name:     "host",
		Type:     schema.FieldTypeUrl,
		Required: true,
		Unique:   true,
		Options: &schema.TextOptions{
			Max: types.Pointer(100),
		},
	})
	form.Schema.AddField(&schema.SchemaField{
		Name:     "enabled",
		Type:     schema.FieldTypeBool,
		Required: true,
	})
	form.Schema.AddField(&schema.SchemaField{
		Name:     "verifySSL",
		Type:     schema.FieldTypeBool,
		Required: true,
	})
	form.Schema.AddField(&schema.SchemaField{
		Name:     "lastactivity",
		Type:     schema.FieldTypeDate,
		Required: true,
	})
	form.Schema.AddField(&schema.SchemaField{
		Name:     "model_id",
		Type:     schema.FieldTypeRelation,
		Required: true,
		Options: &schema.RelationOptions{
			CollectionId:  models_collection.Id,
			CascadeDelete: false,
		},
	})
	if err := form.Submit(); err != nil {
		return err
	}

	return nil

}
func initCollection_Providers(app core.App) error {

	collection := &models.Collection{
		Indexes: types.JsonArray[string]{
			"CREATE UNIQUE INDEX `idx_providers_name` ON `providers` (`name`)",
		},
	}

	form := forms.NewCollectionUpsert(app, collection)
	form.Name = "providers"
	form.Type = models.CollectionTypeBase
	form.ListRule = nil
	form.ViewRule = types.Pointer("@request.auth.id != ''")
	form.CreateRule = types.Pointer("")
	form.UpdateRule = types.Pointer("@request.auth.id != ''")
	form.DeleteRule = nil
	form.Schema.AddField(&schema.SchemaField{
		Name:     "name",
		Type:     schema.FieldTypeText,
		Required: true,
		Unique:   true,
		Options: &schema.TextOptions{
			Max: types.Pointer(100),
		},
	})
	form.Schema.AddField(&schema.SchemaField{
		Name:     "enabled",
		Type:     schema.FieldTypeBool,
		Required: true,
	})

	if err := form.Submit(); err != nil {
		return err
	}

	return nil

}
func initCollection_Models(app core.App, prompts_format_collection *models.Collection, providers_collection *models.Collection) error {

	collection := &models.Collection{
		Indexes: types.JsonArray[string]{
			"CREATE UNIQUE INDEX `idx_models_name` ON `models` (`name`)",
		},
	}

	form := forms.NewCollectionUpsert(app, collection)
	form.Name = "models"
	form.Type = models.CollectionTypeBase
	form.ListRule = nil
	form.ViewRule = types.Pointer("@request.auth.id != ''")
	form.CreateRule = types.Pointer("")
	form.UpdateRule = types.Pointer("@request.auth.id != ''")
	form.DeleteRule = nil
	form.Schema.AddField(&schema.SchemaField{
		Name:     "name",
		Type:     schema.FieldTypeText,
		Required: true,
		Unique:   true,
		Options: &schema.TextOptions{
			Max: types.Pointer(50),
		},
	})
	form.Schema.AddField(&schema.SchemaField{
		Name:     "description",
		Type:     schema.FieldTypeText,
		Required: true,
		Options: &schema.TextOptions{
			Max: types.Pointer(250),
		},
	})
	form.Schema.AddField(&schema.SchemaField{
		Name:     "default_temperature",
		Type:     schema.FieldTypeNumber,
		Required: true,
		Options: &schema.NumberOptions{
			Min: types.Pointer(0.0),
			Max: types.Pointer(1.0),
		},
	})
	form.Schema.AddField(&schema.SchemaField{
		Name:     "top_k",
		Type:     schema.FieldTypeNumber,
		Required: true,
		Options: &schema.NumberOptions{
			Min: types.Pointer(0.0),
			Max: types.Pointer(100.0),
		},
	})
	form.Schema.AddField(&schema.SchemaField{
		Name:     "top_p",
		Type:     schema.FieldTypeNumber,
		Required: true,
		Options: &schema.NumberOptions{
			Min: types.Pointer(0.0),
			Max: types.Pointer(1.0),
		},
	})
	form.Schema.AddField(&schema.SchemaField{
		Name:     "repetition_penality",
		Type:     schema.FieldTypeNumber,
		Required: true,
		Options: &schema.NumberOptions{
			Min: types.Pointer(0.0),
			Max: types.Pointer(2.0),
		},
	})
	form.Schema.AddField(&schema.SchemaField{
		Name:     "max_tokens",
		Type:     schema.FieldTypeNumber,
		Required: true,
		Options: &schema.NumberOptions{
			Min: types.Pointer(1.0),
		},
	})
	form.Schema.AddField(&schema.SchemaField{
		Name:     "truncate",
		Type:     schema.FieldTypeNumber,
		Required: true,
		Options: &schema.NumberOptions{
			Min: types.Pointer(1.0),
		},
	})
	form.Schema.AddField(&schema.SchemaField{
		Name:     "prompt_format_id",
		Type:     schema.FieldTypeRelation,
		Required: true,
		Options: &schema.RelationOptions{
			CollectionId:  prompts_format_collection.Id,
			CascadeDelete: false,
		},
	})
	form.Schema.AddField(&schema.SchemaField{
		Name:     "provider_id",
		Type:     schema.FieldTypeRelation,
		Required: true,
		Options: &schema.RelationOptions{
			CollectionId:  providers_collection.Id,
			CascadeDelete: false,
		},
	})
	if err := form.Submit(); err != nil {
		return err
	}

	return nil

}

func initCollection_Config(app core.App) error {

	collection := &models.Collection{
		Indexes: types.JsonArray[string]{
			"CREATE UNIQUE INDEX `idx_config_name` ON `config` (`name`)",
		},
	}

	form := forms.NewCollectionUpsert(app, collection)
	form.Name = "config"
	form.Type = models.CollectionTypeBase
	form.ListRule = nil
	form.ViewRule = types.Pointer("@request.auth.id != ''")
	form.CreateRule = types.Pointer("")
	form.UpdateRule = types.Pointer("@request.auth.id != ''")
	form.DeleteRule = nil
	form.Schema.AddField(&schema.SchemaField{
		Name:     "name",
		Type:     schema.FieldTypeText,
		Required: true,
		Unique:   true,
		Options: &schema.TextOptions{
			Max: types.Pointer(100),
		},
	})
	form.Schema.AddField(&schema.SchemaField{
		Name:     "value",
		Type:     schema.FieldTypeText,
		Required: true,
		Options: &schema.TextOptions{
			Max: types.Pointer(250),
		},
	})

	if err := form.Submit(); err != nil {
		return err
	}

	return nil

}
