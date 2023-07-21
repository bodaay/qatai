package db

import (
	"context"
	"encoding/json"
	"fmt"

	"go.etcd.io/bbolt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Salt     string `json:"salt"`
	Password string `json:"password"` // This should be hashed & salted in a real-world scenario

}

// BBolt
func (db *BBoltDB) GetUser(id string) (*User, error) {
	var user User

	err := db.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		v := b.Get([]byte(id))

		if v == nil {
			return fmt.Errorf("user not found")
		}

		err := json.Unmarshal(v, &user)
		return err
	})

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (db *BBoltDB) SetUser(user *User) error {
	return db.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("users"))
		if err != nil {
			return err
		}

		data, err := json.Marshal(user)
		if err != nil {
			return err
		}

		return b.Put([]byte(user.ID), data)
	})
}

//Mongo

func (db *MongoDB) GetUser(id string) (*User, error) {
	collection := db.client.Database(db.dbName).Collection("users")
	filter := bson.M{"id": id}

	var user User
	err := collection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (db *MongoDB) SetUser(user *User) error {
	collection := db.client.Database(db.dbName).Collection("users")
	filter := bson.M{"id": user.ID}

	upsert := true
	opts := options.Update().SetUpsert(upsert)

	_, err := collection.UpdateOne(context.Background(), filter, bson.M{"$set": user}, opts)
	return err
}
