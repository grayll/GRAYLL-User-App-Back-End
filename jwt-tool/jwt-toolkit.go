package jwttool

import (
	"crypto/rsa"
	"errors"
	"io/ioutil"
	"log"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

type JwtToolkit struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
	secretKey  []byte
}

func NewJwtFromSecretKey(secretKey []byte) *JwtToolkit {
	return &JwtToolkit{secretKey: secretKey}
}
func NewJwtFromRsaKey(keyPath string) (*JwtToolkit, error) {
	readBytes, err := ioutil.ReadFile(keyPath)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(readBytes)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &JwtToolkit{PrivateKey: privKey, PublicKey: &privKey.PublicKey}, nil
}
func (jwtToolKit *JwtToolkit) GenToken(uid string, exp int) (string, error) {

	claims := jwt.MapClaims{
		"uid": uid,
		"app": "grayll-ua",
		"exp": time.Now().Add(time.Minute * time.Duration(exp)).Unix(),
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// claimInfo := jwt.Claims{uid,
	// 	jwt.StandardClaims{
	// 		ExpiresAt: time.Now().Add(time.Minute * 30).Unix(),
	// 		Issuer:    "grayll-ua",
	// 	}}
	// token := jwt.NewWithClaims(jwt.SigningMethodRS256, claimInfo)
	// tokenString, err := token.SignedString(jwtToolKit.privateKey)
	// if err != nil {
	// 	log.Printf("Token Signing error: %v\n", err)
	// }

	tokenString, err := jwtToken.SignedString(jwtToolKit.PrivateKey)
	if err != nil {
		log.Printf("jwtToken.SignedString: error: %v\n", err)
		return "", err
	}
	return tokenString, nil
}
func (jwtToolKit *JwtToolkit) ParseToken(token string) (jwt.MapClaims, error) {

	jwtToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("There was an error")
		}
		return jwtToolKit.PublicKey, nil
	})

	if err != nil {
		log.Printf("jwtToken.Parse: error: %v\n", err)
		return nil, err
	}
	if !jwtToken.Valid {
		return nil, errors.New("Token is invalid")
	}

	return jwtToken.Claims.(jwt.MapClaims), nil
}
