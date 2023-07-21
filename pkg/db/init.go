package db

import (
	"context"

	"go.etcd.io/bbolt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	client *mongo.Client
}

type BBoltDB struct {
	db *bbolt.DB
}

func InitMongoDB(uri string) (QataiDatabase, error) {
	clientOptions := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, err
	}

	// Ping the database to ensure connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	return newMongoDB(client), nil
}

func InitBBoltDB(path string) (QataiDatabase, error) {
	db, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("users"))
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists([]byte("config"))
		return err
	})

	if err != nil {
		return nil, err
	}

	return newBBoltDB(db), nil
}

func newMongoDB(client *mongo.Client) *MongoDB {
	return &MongoDB{client: client}
}

func newBBoltDB(db *bbolt.DB) *BBoltDB {
	return &BBoltDB{db: db}
}
