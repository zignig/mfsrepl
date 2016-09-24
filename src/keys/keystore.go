package keys

// manage and build key sets
// boltdb store for keys
import (
	"github.com/boltdb/bolt"
	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("keystore")

type KeyStore struct {
	db *bolt.DB
}

func NewKeyStore(path string) (ks *KeyStore, err error) {
	logger.Criticalf("New Key Store %s", path)
	ks = &KeyStore{}
	ks.db, err = bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	return ks, nil

}

func (ks *KeyStore) Close() {
	ks.db.Close()
}

func (ks *KeyStore) Insert(s StoredKey, bucket string) error {
	err := ks.db.Update(func(tx *bolt.Tx) error {
		//bucket, err := tx.CreateBucketIfNotExists([]byte(bucket))
		//if err != nil {
		//	return err
		//}
		//err = bucket.Put(key, value)
		//if err != nil {
		//	return err
		//}
		return nil
	})
	return err
}
