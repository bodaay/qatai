package db

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	QataiDatabaseCommon
	client *mongo.Client
	dbName string
}

func InitNewMongoDB(uriOrPath string, dbName string) (*MongoDB, error) {
	clientOptions := options.Client().ApplyURI(uriOrPath)

	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, err
	}

	// Ping the database to ensure connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	return &MongoDB{client: client, dbName: dbName}, nil
}

func (m *MongoDB) SetValueByKeyName(CollectionBucketName string, Key string, Value string) error {
	collection := m.client.Database(m.dbName).Collection(CollectionBucketName)
	filter := bson.M{"_id": Key}
	update := bson.M{"$set": bson.M{"Value": Value}}
	_, err := collection.UpdateOne(context.Background(), filter, update, options.Update().SetUpsert(true))
	return err
}

func (m *MongoDB) GetValueByKeyName(CollectionBucketName string, Key string) (string, error) {
	collection := m.client.Database(m.dbName).Collection(CollectionBucketName)
	var result struct {
		Value string
	}
	filter := bson.M{"_id": Key}
	err := collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		return "", err
	}
	return result.Value, nil
}

func (db *MongoDB) GetAllRecordForCollectionBucket(CollectionBucketName string) ([]string, error) {
	var results []string
	collection := db.client.Database(db.dbName).Collection(CollectionBucketName)
	cursor, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var doc bson.M
		err = cursor.Decode(&doc)
		if err != nil {
			return nil, err
		}
		for _, value := range doc {
			results = append(results, fmt.Sprintf("%v", value))
		}
	}
	if err = cursor.Err(); err != nil {
		return nil, err
	}
	return results, nil
}
