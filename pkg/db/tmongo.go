package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	client *mongo.Client
	dbName string
}

func InitNewMongoDB(uri string, dbName string) (*MongoDB, error) {
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

	return &MongoDB{
		client: client,
		dbName: dbName,
	}, nil
}

func (m *MongoDB) SetValueByKeyName(CollectionBucketName string, record *QataiDatabaseRecord) error {
	collection := m.client.Database(m.dbName).Collection(CollectionBucketName)
	filter := bson.M{"_id": record.Key}
	update := bson.M{"$set": bson.M{"Value": record.Value}}
	_, err := collection.UpdateOne(context.Background(), filter, update, options.Update().SetUpsert(true))
	return err
}

func (m *MongoDB) GetValueByKeyName(CollectionBucketName string, Key string) (*QataiDatabaseRecord, error) {
	collection := m.client.Database(m.dbName).Collection(CollectionBucketName)
	var result struct {
		Value string `bson:"value"`
	}
	filter := bson.M{"_id": Key}
	err := collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &QataiDatabaseRecord{Key, result.Value}, nil
}

func (db *MongoDB) GetAllRecordForCollectionBucket(CollectionBucketName string) ([]QataiDatabaseRecord, error) {
	var results []QataiDatabaseRecord
	collection := db.client.Database(db.dbName).Collection(CollectionBucketName)
	cursor, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		// Define a map that will hold the document data
		var doc map[string]interface{}
		err = cursor.Decode(&doc)
		if err != nil {
			return nil, err
		}
		// Check if the keys "_id" and "Value" exist in the doc map
		if id, ok := doc["_id"].(string); ok {
			if value, ok := doc["Value"].(string); ok {
				results = append(results, QataiDatabaseRecord{Key: id, Value: value})
			}
		}
	}
	if err = cursor.Err(); err != nil {
		return nil, err
	}
	return results, nil
}
func (db *MongoDB) ClearAllRecordsInCollection(CollectionBucketName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	// call cancel func to avoid context leak
	defer cancel()

	err := db.client.Database(db.dbName).Collection(CollectionBucketName).Drop(ctx)
	if err != nil {
		return err
	}

	return nil
}
