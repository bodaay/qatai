package db

var RequiredCollectionBucket = []string{"config", "users", "models", "chats", "messages", "media", "openai"}

type QataiDatabase interface {
	SetValueByKeyName(CollectionBucketName string, record *QataiDatabaseRecord) error
	GetValueByKeyName(CollectionBucketName string, Key string) (*QataiDatabaseRecord, error)
	GetAllRecordForCollectionBucket(CollectionBucketName string) ([]QataiDatabaseRecord, error) //TODO: I know, wrong, dirty, but I need it anyway and too lazy to create a new a collection for mapping
	ClearAllRecordsInCollection(CollectionBucketName string) error
}

type QataiDatabaseRecord struct {
	Key   string
	Value string
}
