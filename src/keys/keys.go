package keys

// manage and build key sets
// boltdb store for keys
import (
	"github.com/boltdb/bolt"
	"github.com/op/go-logging"

	"crypto/rand"
	"crypto/rsa"
	//	"encoding/base64"
	"crypto/x509"
	"encoding/pem"
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
	var privateKey *rsa.PrivateKey
	var publicKey *rsa.PublicKey
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		logger.Fatal(err)
	}
	publicKey = &privateKey.PublicKey
	privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
	pr := string(pem.EncodeToMemory(privateKeyPEM))
	logger.Criticalf("priv key %v", pr)

	publicKeyDER, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		logger.Fatalf("PUBLIC %v", err)
	}
	publicKeyBlock := pem.Block{
		Type:    "PUBLIC KEY",
		Headers: nil,
		Bytes:   publicKeyDER,
	}
	pb := string(pem.EncodeToMemory(&publicKeyBlock))
	logger.Criticalf("pub key %v", pb)
}
