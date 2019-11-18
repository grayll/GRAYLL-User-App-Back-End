package utils

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"

	"log"

	"crypto/sha256"
	"crypto/subtle"
	"encoding/binary"
	"errors"

	"golang.org/x/crypto/scrypt"
)

const (
	rsaKeySize = 2048
)

func hash(data []byte) []byte {
	s := sha1.Sum(data)
	return s[:]

}
func EncryptItem(publicKey *rsa.PublicKey, item string) string {
	randbytes := make([]byte, 8)
	_, err := rand.Read(randbytes)

	rawbytes := append(randbytes, []byte(item)...)
	code, err := PublicEncrypt(publicKey, rawbytes)
	if err != nil {
		return ""
	}
	encodeStr := base64.URLEncoding.EncodeToString(code)
	log.Println("EncryptItem: encodestr: ", encodeStr)
	return encodeStr
}
func DecryptItem(privKey *rsa.PrivateKey, item string) string {
	encodeBytes, err := base64.URLEncoding.DecodeString(item)
	if err != nil {
		log.Println("DecryptItem: err: ", err)
		return ""
	}
	decodeBytes, err := PrivateDecrypt(privKey, encodeBytes)
	if err != nil {
		log.Println("PrivateDecrypt: err: ", err)
		return ""
	}
	itemDecodeBytes := decodeBytes[8:]
	return string(itemDecodeBytes)
}
func GenerateKey() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	pri, err := rsa.GenerateKey(rand.Reader, rsaKeySize)
	if err != nil {
		return nil, nil, err
	}
	return pri, &pri.PublicKey, nil
}

func GenerateKeyBytes() (privateBytes, publicBytes []byte, err error) {
	pri, pub, err := GenerateKey()
	if err != nil {
		return nil, nil, err
	}
	priBytes, err := x509.MarshalPKCS8PrivateKey(pri)
	if err != nil {
		return nil, nil, err
	}
	pubBytes := x509.MarshalPKCS1PublicKey(pub)
	return priBytes, pubBytes, nil
}

func GenerateKey64() (pri64, pub64 string, err error) {
	pri, pub, err := GenerateKeyBytes()
	if err != nil {
		return "", "", nil
	}
	return base64.StdEncoding.EncodeToString(pri),
		base64.StdEncoding.EncodeToString(pub),
		nil
}

func PublicKeyFrom(key []byte) (*rsa.PublicKey, error) {
	pubInterface, err := x509.ParsePKIXPublicKey(key)
	if err != nil {
		return nil, err
	}
	pub, ok := pubInterface.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("invalid public key")
	}
	return pub, nil
}

func PublicKeyFrom64(key string) (*rsa.PublicKey, error) {
	b, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, err
	}
	return PublicKeyFrom(b)
}

func PrivateKeyFrom(key []byte) (*rsa.PrivateKey, error) {
	pri, err := x509.ParsePKCS8PrivateKey(key)
	if err != nil {
		return nil, err
	}
	p, ok := pri.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("invalid private key")
	}
	return p, nil
}

func PrivateKeyFrom64(key string) (*rsa.PrivateKey, error) {
	b, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, err
	}
	return PrivateKeyFrom(b)
}

func PublicEncrypt(key *rsa.PublicKey, data []byte) ([]byte, error) {
	return rsa.EncryptPKCS1v15(rand.Reader, key, data)
}

func PublicSign(key *rsa.PublicKey, data []byte) ([]byte, error) {
	return PublicEncrypt(key, hash(data))
}

func PublicVerify(key *rsa.PublicKey, sign, data []byte) error {
	return rsa.VerifyPKCS1v15(key, crypto.SHA1, hash(data), sign)
}

func PrivateDecrypt(key *rsa.PrivateKey, data []byte) ([]byte, error) {
	return rsa.DecryptPKCS1v15(rand.Reader, key, data)
}

func PrivateSign(key *rsa.PrivateKey, data []byte) ([]byte, error) {
	return rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA1, hash(data))
}

func PrivateVerify(key *rsa.PrivateKey, sign, data []byte) error {
	h, err := PrivateDecrypt(key, sign)
	if err != nil {
		return err
	}
	if !bytes.Equal(h, hash(data)) {
		return rsa.ErrVerification
	}
	return nil
}

func VerifyPassword(hashPass, password string) bool {
	decodeBytes, err := base64.StdEncoding.DecodeString(hashPass)
	if err != nil {
		log.Println("VerifyPassword: DecodeString error: ", err)
		return false
	}
	salt := decodeBytes[:44]
	hash := decodeBytes[44:]

	derivedKey, err := scrypt.Key([]byte(password), salt, 16384, 8, 1, 16)
	if bytes.Compare(hash, derivedKey) == 0 {
		return true
	}
	return false
}

const (
	N                = 16384
	r                = 8
	p                = 1
	metadataLenBytes = 60
	saltLenBytes     = 16
)

// DerivePassphrase returns a keylenBytes+60 bytes of derived text
// from the input passphrase.
// It runs the scrypt function for this.
func DerivePassphrase(passphrase string, keylenBytes int) (string, error) {
	// Generate salt
	salt, err := generateSalt()
	if err != nil {
		return "", err
	}

	// Generate key
	key, err := scrypt.Key([]byte(passphrase),
		salt,
		N, // Must be a power of 2 greater than 1
		r,
		p, // r*p must be < 2^30
		keylenBytes)
	if err != nil {
		return "", err
	}

	// Appending the salt
	key = append(key, salt...)

	// Encoding the params to be stored
	buf := &bytes.Buffer{}
	for _, elem := range [3]int{N, r, p} {
		err = binary.Write(buf, binary.LittleEndian, int32(elem))
		if err != nil {
			return "", err
		}
	}
	key = append(key, buf.Bytes()...)

	// appending the sha-256 of the entire header at the end
	hashDigest := sha256.New()
	_, err = hashDigest.Write(key)
	if err != nil {
		return "", err
	}
	hash := hashDigest.Sum(nil)
	key = append(key, hash...)
	keyStr := base64.StdEncoding.EncodeToString(key)
	return keyStr, nil
}

// VerifyPassphrase takes the passphrase and the targetKey to match against.
// And returns a boolean result whether it matched or not
func VerifyPassphrase(passphrase string, targetKeyStr string) (bool, error) {
	targetKey, err := base64.StdEncoding.DecodeString(targetKeyStr)
	if err != nil {
		log.Println("VerifyPassphrase: DecodeString error: ", err)
		return false, err
	}
	keylenBytes := len(targetKey) - metadataLenBytes
	if keylenBytes < 1 {
		return false, errors.New("Invalid targetKey length")
	}
	// Get the master_key
	targetMasterKey := targetKey[:keylenBytes]
	// Get the salt
	salt := targetKey[keylenBytes : keylenBytes+saltLenBytes]
	// Get the params
	var N, r, p int32
	paramsStartIndex := keylenBytes + saltLenBytes

	err = binary.Read(bytes.NewReader(targetKey[paramsStartIndex:paramsStartIndex+4]), // 4 bytes for N
		binary.LittleEndian,
		&N)
	if err != nil {
		return false, err
	}

	err = binary.Read(bytes.NewReader(targetKey[paramsStartIndex+4:paramsStartIndex+8]), // 4 bytes for r
		binary.LittleEndian,
		&r)
	if err != nil {
		return false, err
	}

	err = binary.Read(bytes.NewReader(targetKey[paramsStartIndex+8:paramsStartIndex+12]), // 4 bytes for p
		binary.LittleEndian,
		&p)
	if err != nil {
		return false, err
	}
	sourceMasterKey, err := scrypt.Key([]byte(passphrase),
		salt,
		int(N), // Must be a power of 2 greater than 1
		int(r),
		int(p), // r*p must be < 2^30
		keylenBytes)
	if err != nil {
		return false, err
	}

	targetHash := targetKey[paramsStartIndex+12:]
	// Doing the sha-256 checksum at the last because we want the attacker
	// to spend as much time possible cracking
	hashDigest := sha256.New()
	_, err = hashDigest.Write(targetKey[:paramsStartIndex+12])
	if err != nil {
		return false, err
	}
	sourceHash := hashDigest.Sum(nil)

	// ConstantTimeCompare returns ints. Converting it to bool
	keyComp := subtle.ConstantTimeCompare(sourceMasterKey, targetMasterKey) != 0
	hashComp := subtle.ConstantTimeCompare(targetHash, sourceHash) != 0
	result := keyComp && hashComp
	return result, nil
}

func generateSalt() ([]byte, error) {
	salt := make([]byte, saltLenBytes)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}
