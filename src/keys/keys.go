package keys

// manage and build key sets
// boltdb store for keys
import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/gob"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
)

const KeySize = 1024
const FingerPrintSize = 32

var (
	ErrNoPrivate  = errors.New("No Private Key")
	ErrBadPem     = errors.New("Bad Pem Block")
	ErrBadPemType = errors.New("Bad Pem Type")
)

type StoredKey struct {
	HavePrivate bool
	Private     string
	Public      string
}

// Key with finger print
type DistKey struct {
	PublicKey   string //pem format
	Fingerprint string
}

// Public Key signed with itself for mesh gossip
type SignedKey struct {
	Key DistKey
	Sig string
}

// Trucated finger SHA256 of the public key
func (sk *StoredKey) Fingerprint() (fp string) {
	data := sha256.Sum256([]byte(sk.Public))
	fp = hex.EncodeToString(data[:])
	return fp[:FingerPrintSize]
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

// takes a stored key and makes a distribution key
func (sk *StoredKey) MakeSigned() (sig *SignedKey, err error) {
	if sk.HavePrivate == false {
		return nil, ErrNoPrivate
	}
	block, _ := pem.Decode([]byte(sk.Private))
	if block == nil {
		return nil, ErrBadPem
	}
	fmt.Println(block.Type)
	if block.Type != "RSA PRIVATE KEY" {
		return nil, ErrBadPemType
	}
	fmt.Println(block.Headers)
	//priv := block.Bytes

	//hash := crypto.SHA1
	//h := hash.New()
	//h.Write(data)
	//hashed := h.Sum(nil)
	//signature, err := rsa.SignPKCS1v15(rand.Reader, s.Key, hash, hashed)
	//if err != nil {
	//	return nil, err
	//}

	return nil, nil
}

func (ks *KeyStore) NewLocalKey() (lc *StoredKey, err error) {
	var privateKey *rsa.PrivateKey
	var publicKey *rsa.PublicKey
	privateKey, err = rsa.GenerateKey(rand.Reader, KeySize)
	if err != nil {
		logger.Critical(err)
		return nil, err
	}
	publicKey = &privateKey.PublicKey
	//TODO, add some header stuff
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
	}
	return lc, nil
}
