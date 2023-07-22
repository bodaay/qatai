package db

type Config struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

const collection_bucket_name = "config"

// BBolt
func (db *QataiDatabaseCommon) GetConfig(key string) (*Config, error) {
	var config Config
	v, err := db.GetValueByKeyName(collection_bucket_name, key)
	if err != nil {
		return nil, err
	}
	config.Value = v

	return &config, nil
}

func (db *QataiDatabaseCommon) SetConfig(config *Config) error {

	err := db.SetValueByKeyName(collection_bucket_name, config.Key, config.Value)
	if err != nil {
		return err
	}
	return nil
}
