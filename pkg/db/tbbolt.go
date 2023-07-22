package db

import (
	"fmt"

	bolt "go.etcd.io/bbolt"
)

type BBoltDB struct {
	QataiDatabaseCommon
	db     *bolt.DB
	dbPath string
}

func InitNewBoltDB(uriOrPath string, dbName string) (*BBoltDB, error) { // add all BBolt
	db, err := bolt.Open(uriOrPath, 0600, nil)
	if err != nil {
		return nil, err
	}
	err = createCollectionBucket(db, RequiredCollectionBucket)
	if err != nil {
		return nil, err
	}

	return &BBoltDB{db: db, dbPath: uriOrPath}, nil
}

func createCollectionBucket(db *bolt.DB, names []string) error {
	for _, bbucket := range names {
		err := db.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte(bbucket))
			return err
		})
		if err != nil {
			return err
		}
	}
	return nil
}
func (b *BBoltDB) SetValueByKeyName(CollectionBucketName string, Key string, Value string) {
	b.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(CollectionBucketName))
		if err != nil {
			return err
		}

		err = bucket.Put([]byte(Key), []byte(Value))
		return err
	})
}

func (b *BBoltDB) GetValueByKeyName(CollectionBucketName string, Key string) string {
	var value string
	b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(CollectionBucketName))
		if bucket != nil {
			value = string(bucket.Get([]byte(Key)))
		}
		return nil
	})
	return value
}

func (db *BBoltDB) GetAllRecordForCollectionBucket(CollectionBucketName string) ([]string, error) {
	var results []string
	err := db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(CollectionBucketName))
		if b == nil {
			return fmt.Errorf("no bucket named %s found", CollectionBucketName)
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			results = append(results, fmt.Sprintf("%s: %s", k, v))
		}
		return nil
	})
	if err != nil {
		return results, err
	}
	return results, nil
}
