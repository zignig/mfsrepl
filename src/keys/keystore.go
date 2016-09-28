package keys

// manage and build key sets
// boltdb store for keys
import (
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("keystore")

var (
	ErrNoKey = errors.New("Key does not exist")
)

type KeyStore struct {
	db      *bolt.DB
	private *rsa.PrivateKey
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

func (ks *KeyStore) ListKeys(bucket string) (err error) {
	fmt.Println("Bucket ", bucket)
	err = ks.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucket))
		c := bucket.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			fmt.Printf("key->%s\n", k)
		}
		return nil
	})
	return err
}

func (ks *KeyStore) GetPublic(fp, bucket string) (sigK *SignedKey, err error) {
	err = ks.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucket))
		data := bucket.Get([]byte(fp))
		if data == nil {
			return ErrNoKey
		}
		sigK, err := DecodeSignedKey(data)
		if err != nil {
			return err
		}
		err = sigK.Check()
		if err != nil {
			return err
		}
		return nil
	})
	return sigK, err
}

func (ks *KeyStore) PutPublic(sigK *SignedKey, bucket string) error {
	err := ks.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}
		key, err := sigK.GetFingerPrint()
		if err != nil {
			return err
		}
		data, err := sigK.Encode()
		if err != nil {
			return err
		}
		err = bucket.Put([]byte(key), data)
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func (ks *KeyStore) Insert(s *StoredKey, bucket string) error {
	err := ks.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}
		data, err := s.Encode()
		if err != nil {
			return err
		}
		// truncated sha of the public key
		fp := s.FingerPrint()
		err = bucket.Put([]byte(fp), data)
		if err != nil {
			return err
		}
		return nil
	})
	return err
}
