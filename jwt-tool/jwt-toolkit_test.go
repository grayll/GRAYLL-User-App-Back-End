package jwttool

import (
	"log"
	"testing"
)

// Gen rsa from cmd
// openssl genrsa -out grua.key 2048
func TestGenToken(t *testing.T) {
	jwtToolkit, err := NewJwtFromRsaKey("grua.pem")
	if err != nil {
		t.Fail()
	}
	log.Println(jwtToolkit.PrivateKey)
	tokenString, err := jwtToolkit.GenToken("iur454efdjfdfdfdf", 30)
	if err != nil {
		t.Errorf("Failed with error %v", err)
	}
	log.Println("GenToken: tokenString:", tokenString)

	claims, err := jwtToolkit.ParseToken(tokenString)
	if err != nil {
		t.Errorf("ParseToken: Failed with error %v", err)
	}

	log.Println("Uid: ", claims["uid"])
	log.Println("exp: ", claims["exp"])
}
