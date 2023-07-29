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
	app.OnBeforeServe().Add(
		func(e *core.ServeEvent) error {
			if _, err := app.Dao().FindCollectionByNameOrId("models"); err != nil {

				collection := &models.Collection{}

				form := forms.NewCollectionUpsert(e.App, collection)
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
					// Options: &schema.RelationOptions{
					// 	MaxSelect:     types.Pointer(1),
					// 	CollectionId:  "ae40239d2bc4477",
					// 	CascadeDelete: true,
					// },
				})
				if err := form.Submit(); err != nil {
					return err
				}
			}
			return nil
		},
	)

	return nil
}
