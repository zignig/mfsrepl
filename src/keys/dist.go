package keys

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
)

var (
	ErrFingerPrintSize  = errors.New("Incorrect Finger Print Size")
	ErrFingerPrintCheck = errors.New("Finger Print does not match")
)

func FingerPrint(data string) (fp string) {
	shaData := sha256.Sum256([]byte(data))
	fp = hex.EncodeToString(shaData[:])
	return fp[:FingerPrintSize]
}
func (sigK *SignedKey) Encode() (b []byte, err error) {
	b, err = json.MarshalIndent(sigK, " ", " ")
	return
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
