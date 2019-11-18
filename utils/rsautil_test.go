package utils

import (
	"fmt"
	"testing"
)

func TestScrypt(t *testing.T) {
	passphrase := "Hello there this is a sample passphrase"

	key, err := DerivePassphrase(passphrase, 32)
	if err != nil {
		fmt.Errorf("Error returned: %s\n", err)
	}

	fmt.Printf("Key returned - %v\n", key)
	var result bool

	result, err = VerifyPassphrase(passphrase, key)
	if err != nil {
		fmt.Printf("Error returned: %s\n", err)
	}
	if !result {
		fmt.Printf("Passphrase did not match\n")
	} else {
		fmt.Printf("Passphrase matched successfully\n")
	}
}
