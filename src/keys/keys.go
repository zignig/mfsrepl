package keys

// manage and build key sets
// boltdb store for keys
import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
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

// Trucated finger SHA256 of the public key
func (sk *StoredKey) FingerPrint() (fp string) {
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

func GetPrivateFromPem(data string) (key *rsa.PrivateKey, err error) {
	block, _ := pem.Decode([]byte(data))
	if block == nil {
		return nil, ErrBadPem
	}
	if block.Type != "RSA PRIVATE KEY" {
		return nil, ErrBadPemType
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	err = privateKey.Validate()
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func GetPublicFromPem(data string) (key *rsa.PublicKey, err error) {
	block, _ := pem.Decode([]byte(data))
	if block == nil {
		return nil, ErrBadPem
	}
	if block.Type != "PUBLIC KEY" {
		fmt.Println(block.Type)
		return nil, ErrBadPemType
	}
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return publicKey.(*rsa.PublicKey), nil
}

// takes a stored key and makes a distribution key
func (sk *StoredKey) MakeSigned() (sig *SignedKey, err error) {
	if sk.HavePrivate == false {
		return nil, ErrNoPrivate
	}
	//Get Private Key
	privKey, err := GetPrivateFromPem(sk.Private)
	if err != nil {
		return nil, err
	}
	dk := &DistKey{
		PublicKey:   sk.Public,
		FingerPrint: sk.FingerPrint(),
	}
	jsonData, err := json.MarshalIndent(dk, " ", " ")
	if err != nil {
		return nil, err
	}
	hash := crypto.SHA256
	h := hash.New()
	h.Write(jsonData)
	hashed := h.Sum(nil)
	var opts rsa.PSSOptions
	opts.SaltLength = rsa.PSSSaltLengthAuto
	signature, err := rsa.SignPSS(rand.Reader, privKey, hash, hashed, &opts)
	if err != nil {
		return nil, err
	}

	sig = &SignedKey{
		Data:      jsonData,
		Signature: hex.EncodeToString(signature[:]),
	}
	return sig, nil
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

func (lc *StoredKey) Save(path string) (err error) {
	err = os.Mkdir(path, 0700)
	enc, err := json.MarshalIndent(lc, "", " ")
	if err != nil {
		return err
	}
	fp := lc.FingerPrint()
	err = ioutil.WriteFile(path+string(os.PathSeparator)+fp+".key", enc, 0600)
	if err != nil {
		return err

	}
	return nil
}
