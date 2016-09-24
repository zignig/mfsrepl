package keys

// manage and build key sets
// boltdb store for keys
import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/gob"
	"encoding/pem"
	"time"
)

type StoredKey struct {
	HavePrivate bool
	Private     string
	Public      string
	Date        time.Time
}

func (sk *StoredKey) Encode() (data []byte, err error) {
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	err = enc.Encode(sk)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func Decode(data []byte) (sk *StoredKey, err error) {
	b := bytes.NewBuffer(data)
	dec := gob.NewDecoder(b)
	err = dec.Decode(&sk)
	if err != nil {
		return nil, err
	}
	return sk, err
}

func (ks *KeyStore) NewLocalKey() (lc *StoredKey, err error) {
	var privateKey *rsa.PrivateKey
	var publicKey *rsa.PublicKey
	privateKey, err = rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		logger.Critical(err)
		return nil, err
	}
	publicKey = &privateKey.PublicKey

	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	pr := string(pem.EncodeToMemory(privateKeyPEM))

	publicKeyDER, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		logger.Criticalf("PUBLIC %v", err)
		return nil, err
	}
	publicKeyBlock := pem.Block{
		Type:    "PUBLIC KEY",
		Headers: nil,
		Bytes:   publicKeyDER,
	}
	pb := string(pem.EncodeToMemory(&publicKeyBlock))

	lc = &StoredKey{
		HavePrivate: true,
		Private:     pr,
		Public:      pb,
		Date:        time.Now(),
	}
	return lc, nil
}
