package db

import (
	"fmt"

	bolt "go.etcd.io/bbolt"
)

type BBoltDB struct {
	db     *bolt.DB
	dbPath string
}

func InitNewBoltDB(path string) (*BBoltDB, error) { // add all BBolt
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}
	err = createCollectionBucket(db, RequiredCollectionBucket)
	if err != nil {
		return nil, err
	}

	return &BBoltDB{
		db:     db,
		dbPath: path,
	}, nil
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
func (b *BBoltDB) SetValueByKeyName(CollectionBucketName string, record *QataiDatabaseRecord) error {
	err := b.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(CollectionBucketName))
		if err != nil {
			return err
		}

		err = bucket.Put([]byte(record.Key), []byte(record.Value))
		return err
	})
	return err
}

func (b *BBoltDB) GetValueByKeyName(CollectionBucketName string, Key string) (*QataiDatabaseRecord, error) {
	var value string
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(CollectionBucketName))
		if bucket != nil {
			value = string(bucket.Get([]byte(Key)))
			return nil
		}
		return fmt.Errorf("bucket does not exists")
	})
	if err != nil {
		return nil, err
	}
	return &QataiDatabaseRecord{Key, value}, nil
}

func (db *BBoltDB) GetAllRecordForCollectionBucket(CollectionBucketName string) ([]QataiDatabaseRecord, error) {
	var results []QataiDatabaseRecord
	err := db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(CollectionBucketName))
		if b == nil {
			return fmt.Errorf("no bucket named %s found", CollectionBucketName)
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			results = append(results, QataiDatabaseRecord{Key: string(k), Value: string(v)})
		}
		return nil
	})
	if err != nil {
		return results, err
	}
	return results, nil
}

func (db *BBoltDB) ClearAllRecordsInCollection(CollectionBucketName string) error {
	err := db.db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte(CollectionBucketName))
	})
	if err != nil {
		return err
	}
	//recreate collections in case of bolt, since its needed
	return createCollectionBucket(db.db, RequiredCollectionBucket)
}
