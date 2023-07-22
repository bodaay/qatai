package db

type Config struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

const collection_bucket_name = "config"

// BBolt
func GetConfig(db QataiDatabase, key string) (*Config, error) {
	var config Config
	v, err := db.GetValueByKeyName(collection_bucket_name, key)
	if err != nil {
		return nil, err
	}
	config.Value = v.Value
	config.Key = v.Key
	return &config, nil
}

func SetConfig(db QataiDatabase, config *Config) error {
	record := QataiDatabaseRecord(*config)
	err := db.SetValueByKeyName(collection_bucket_name, &record)
	if err != nil {
		return err
	}
	return nil
}

func GetAllConfig(db QataiDatabase) ([]Config, error) {
	var configs []Config
	values, err := db.GetAllRecordForCollectionBucket(collection_bucket_name)
	if err != nil {
		return nil, err
	}
	for _, v := range values {
		configs = append(configs, Config(v))
	}

	return configs, nil
}

func ClearAllConfig(db QataiDatabase) error {
	return db.ClearAllRecordsInCollection(collection_bucket_name)
}
