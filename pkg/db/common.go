package db

import (
	"go.etcd.io/bbolt"
	"go.mongodb.org/mongo-driver/mongo"
)

type Database interface {
	GetConfig(key string) (*Config, error)
	SetConfig(config *Config) error
}

type MongoDB struct {
	client *mongo.Client
}

type BBoltDB struct {
	db *bbolt.DB
}

func NewMongoDB(client *mongo.Client) *MongoDB {
	return &MongoDB{client: client}
}

func NewBBoltDB(db *bbolt.DB) *BBoltDB {
	return &BBoltDB{db: db}
}
