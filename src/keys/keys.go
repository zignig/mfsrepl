package keys

// manage and build key sets
// boltdb store for keys
import (
	"github.com/boltdb/bolt"
	"github.com/op/go-logging"

	"crypto/rand"
	"crypto/rsa"

	"encoding/base64"
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

func (ks KeyStore) Close() {
	ks.db.Close()
}

func (ks KeyStore) NewKey() {
	privateKey := new(ecdsa.PrivateKey)
	privateKey, err := ecdsa.GenerateKey(pubkeyCurve, rand.Reader)
	if err != nil {
		logger.Fatal(err)
	}
	var pubkey ecdsa.PublicKey
	pubkey = privateKey.PublicKey
	logger.Criticalf("public key: %v", pubkey)
	logger.Criticalf("public key: %v", privateKey)
	str := base64.StdEncoding.EncodeToString(pubkey)
	logger.Criticalf(str)

}
