package keys

// manage and build key sets
// boltdb store for keys
import (
	"crypto/rsa"
	"errors"
	//"fmt"
	"github.com/boltdb/bolt"
	"sync"
)

var (
	ErrNoKey = errors.New("Key does not exist")
)

type KeyStore struct {
	db      *bolt.DB
	private *rsa.PrivateKey

	mapLock sync.Mutex
	keySets map[string]*state // reuse state for key cache
}

func NewKeyStore(path string) (ks *KeyStore, err error) {
	logger.Criticalf("New Key Store %s", path)
	ks = &KeyStore{}
	ks.db, err = bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}
	ks.makeBucket("private")
	ks.makeBucket("public")
	ks.keySets = make(map[string]*state)
	return ks, nil
}

func (ks *KeyStore) makeBucket(bucket string) (err error) {
	err = ks.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		return err
	})
	return err
}

func (ks *KeyStore) Close() {
	ks.db.Close()
}

func (ks *KeyStore) TryInsert(sigK *SignedKey, bucket string) (err error) {
	logger.Critical(ks)
	err = sigK.Check()
	if err != nil {
		return err
	}
	err = ks.PutPublic(sigK, bucket)
	if err != nil {
		return err
	}
	return nil
}

func (ks *KeyStore) HaveKey(fp string, bucket string) (have bool) {
	_, have = ks.CacheKey(fp, bucket)
	return have
}

func (ks *KeyStore) CacheKey(fp string, bucket string) (sigK *SignedKey, have bool) {
	ks.mapLock.Lock()
	defer ks.mapLock.Unlock()
	// a state set for each bucket name
	var keySet *state
	// does the bucket exist
	_, ok := ks.keySets[bucket]
	if ok {
		keySet = ks.keySets[bucket]
	} else {
		// does not exist create the bucket
		ks.keySets[bucket] = newState()
		keySet = ks.keySets[bucket]
	}
	// check for the key in the bucket
	keySet.mtx.Lock()
	defer keySet.mtx.Unlock()

	sigK, ok = keySet.set[fp]
	// do we have the key
	if ok {
		return sigK, true
	}
	// if not load it
	sigK, err := ks.GetPublic(fp, bucket)
	if err != nil {
		return nil, false
	}
	// put the key into the map
	keySet.set[fp] = sigK
	return sigK, true
}

func (ks *KeyStore) ListKeys(bucket string) (items []string, err error) {
	items = make([]string, 0)
	err = ks.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucket))
		c := bucket.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			//fmt.Printf("key->%s\n", k)
			items = append(items, string(k))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return items, err
}

func (ks *KeyStore) GetPublic(fp, bucket string) (sigK *SignedKey, err error) {
	//logger.Criticalf("%s", fp)
	err = ks.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucket))
		data := bucket.Get([]byte(fp))
		if data == nil {
			return ErrNoKey
		}
		sigK, err = DecodeSignedKey(data)
		if err != nil {
			return err
		}
		//Make sure the key is valid
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
