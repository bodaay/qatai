package db

import (
	"context"
	"encoding/json"
	"fmt"

	"go.etcd.io/bbolt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// BBolt
func (db *BBoltDB) GetConfig(key string) (*Config, error) {
	var config Config

	err := db.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		v := b.Get([]byte(key))

		if v == nil {
			return fmt.Errorf("config not found")
		}

		err := json.Unmarshal(v, &config)
		return err
	})

	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (db *BBoltDB) SetConfig(config *Config) error {
	return db.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("config"))
		if err != nil {
			return err
		}

		data, err := json.Marshal(config)
		if err != nil {
			return err
		}

		return b.Put([]byte(config.Key), data)
	})
}

// Mongo
func (db *MongoDB) GetConfig(key string) (*Config, error) {
	collection := db.client.Database("test").Collection("config")
	filter := bson.M{"key": key}

	var config Config
	err := collection.FindOne(context.Background(), filter).Decode(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (db *MongoDB) SetConfig(config *Config) error {
	collection := db.client.Database("test").Collection("config")
	filter := bson.M{"key": config.Key}

	upsert := true
	opts := options.Update().SetUpsert(upsert)

	_, err := collection.UpdateOne(context.Background(), filter, bson.M{"$set": config}, opts)
	return err
}
