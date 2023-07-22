package db

var RequiredCollectionBucket = []string{"config", "users", "models", "chats", "messages", "media", "openai"}

type QataiDatabase interface {
	SetValueByKeyName(CollectionBucketName string, Key string, Value string) error
	GetValueByKeyName(CollectionBucketName string, Key string) (string, error)
	GetAllRecordForCollectionBucket(CollectionBucketName string) ([]string, error) //TODO: I know, wrong, dirty, but I need it anyway and too lazy to create a new a collection for mapping
}

type QataiDatabaseCommon struct {
	QataiDatabase
}
