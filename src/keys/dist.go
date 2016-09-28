package keys

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
)

var (
	ErrFingerPrintSize  = errors.New("Incorrect Finger Print Size")
	ErrFingerPrintCheck = errors.New("Finger Print does not match")
)

// Key with finger print
type DistKey struct {
	PublicKey   string //pem format
	FingerPrint string
}

// Public Key signed with itself for mesh gossip
type SignedKey struct {
	Data      json.RawMessage
	Signature string
}

func FingerPrint(data string) (fp string) {
	shaData := sha256.Sum256([]byte(data))
	fp = hex.EncodeToString(shaData[:])
	return fp[:FingerPrintSize]
}

func (sigK *SignedKey) Encode() (data []byte, err error) {
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	err = enc.Encode(sigK)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func DecodeSignedKey(data []byte) (sigK *SignedKey, err error) {
	b := bytes.NewBuffer(data)
	dec := gob.NewDecoder(b)
	err = dec.Decode(&sigK)
	if err != nil {
		return nil, err
	}
	return sigK, err
}

func (sigK *SignedKey) GetFingerPrint() (fp string, err error) {
	dk, err := sigK.GetDistKey()
	if err != nil {
		return "", err
	}
	return dk.FingerPrint, nil
}

func (sigK *SignedKey) GetDistKey() (dk *DistKey, err error) {
	data := sigK.Data
	dk = &DistKey{}
	// Unmarshall the json
	err = json.Unmarshal(data, dk)
	if err != nil {
		return nil, err
	}
	return dk, nil
}

func (sigK *SignedKey) Check() (err error) {
	signature, err := hex.DecodeString(sigK.Signature)
	if err != nil {
		return err
	}
	data := sigK.Data
	dk := &DistKey{}
	// Unmarshall the json
	err = json.Unmarshal(data, dk)
	if err != nil {
		return err
	}
	// Check the finger print size
	if len(dk.FingerPrint) != FingerPrintSize {
		return ErrFingerPrintSize
	}
	fp := FingerPrint(dk.PublicKey)
	// Check the finger prints aginst eachother
	if strings.Compare(fp, dk.FingerPrint) != 0 {
		return ErrFingerPrintCheck
	}
	// Get The Public Key
	publicKey, err := GetPublicFromPem(dk.PublicKey)
	if err != nil {
		return err
	}
	// Check the Signature
	var opts rsa.PSSOptions
	opts.SaltLength = rsa.PSSSaltLengthAuto
	newhash := crypto.SHA256
	pssh := newhash.New()
	// use the raw data before unmarshalling
	pssh.Write(data)
	hashed := pssh.Sum(nil)
	err = rsa.VerifyPSS(publicKey, crypto.SHA256, hashed, signature, &opts)
	if err != nil {
		return err
	}
	return nil
}
