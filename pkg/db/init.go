package db

import (
	"context"

	"go.etcd.io/bbolt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var requiredBuckets = []string{"config", "users", "models", "chats"}

type MongoDB struct {
	client *mongo.Client
	dbName string
}

type BBoltDB struct {
	db     *bbolt.DB
	dbPath string
}

func InitMongoDB(uri string, dbName string) (QataiDatabase, error) {
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

	ndb := newMongoDB(client)
	ndb.dbName = dbName
	return ndb, nil
}

func InitBBoltDB(dbPath string) (QataiDatabase, error) { // add all BBolt
	db, err := bbolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil, err
	}

	for _, bbucket := range requiredBuckets {
		err = db.Update(func(tx *bbolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte(bbucket))
			return err
		})
		if err != nil {
			return nil, err
		}
	}

	ndb := newBBoltDB(db)
	ndb.dbPath = dbPath
	return ndb, nil
}

func newMongoDB(client *mongo.Client) *MongoDB {
	return &MongoDB{client: client}
}

func newBBoltDB(db *bbolt.DB) *BBoltDB {
	return &BBoltDB{db: db}
}
