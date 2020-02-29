package main

//"testing"

//"github.com/stellar/go/assets"

//build "github.com/stellar/go/txnbuild"
//"github.com/stellar/go/build"

//"log"
//stellar "github.com/huyntsgs/stellar-service"
//"github.com/huyntsgs/stellar-service/assets"
//"github.com/stellar/go/clients/horizonclient"

// issuerSeed := "SCH3R5ONBEKEBD22FNFCCQRAAJHC7X5L7XMETEYUJ7HQQDMSJL3OMCRE" => GRXT
// issuerAddress = GAKXWUADYNO67NQ6ET7PT2DSLE5QGGDTNZZRESXCWWYYA2UCLOYT7AKR
// recipientSeed := "SBTGPMJP2YOEMAD2C3NUXRPR55W6TACNIBIHASLPM6ZRIM5625NDNB4P"
// GAHNPX6MDZRY4ZGEDRHNN3MQ4GHY26UXI3SSO3QHZQGFTHCR7WDHRD6T
// https://stellar.stackexchange.com/questions/2616/how-to-create-trustline-between-issuers-account-and-receivers-accounts-dynamica/2617

func main() {
	//issuerSeed := "SCH3R5ONBEKEBD22FNFCCQRAAJHC7X5L7XMETEYUJ7HQQDMSJL3OMCRE"
	//recipientSeed := "SDJBH532IDUAQT2Z3QGU3W5QIWNC4XDTKDVKRBNGZ6IKUL5NEPNUX2Y7"

	//recipientSeed := "SBTGPMJP2YOEMAD2C3NUXRPR55W6TACNIBIHASLPM6ZRIM5625NDNB4P"
	//dest := "GAP2OSRNSLTCFXBFC26QZ3IYS72OAYOT5UBGMRHRKWGUIQ5X5LMN3KLO"

	// privateKey, publicKey, err := webpush.GenerateVAPIDKeys()
	// if err != nil {
	// 	// TODO: Handle error
	// }

	// Keys for accounts to issue and receive the new asset
	// issuer, err := keypair.Parse(issuerSeed)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// recipient, err := keypair.Parse(recipientSeed)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// log.Println("recipient address:", recipient.Address())
	// //log.Println("recipient address:", recipient.Seed())
	// log.Println("issuer address:", issuer.Address())

	//stellar.SetupParams(float64(1000), false)

	// bl, err := stellar.GetNativeBalance(issuer.Address())
	// if err != nil {
	// 	log.Println("GetNativeBalance:err", err)
	// } else {
	// 	log.Println("GetNativeBalance:balance issuer:", bl)
	// }

	// err = stellar.GetXLM(issuer.Address())
	// if err != nil {
	// 	log.Println("GetXLM issuer:", err)
	// 	return
	// } else {
	// 	log.Println("Send XLM to issuer:", issuer.Address())
	// }
	//grx := assets.CreateAsset("GRXT", issuer.Address())

	//asset := assets.Asset{Code: "GRXT", IssuerAddress: issuer.Address()}

	// bl, err = stellar.GetNativeBalance(recipient.Address())
	// if err != nil {
	// 	log.Println("GetNativeBalance:receipt:err", err)
	// } else {
	// 	log.Println("GetNativeBalance:receipt:balance", bl)
	// }
	// bl, err = stellar.GetAssetBalance(issuer.Address(), "GRXT")
	// log.Println("GetAssetBalance:issuer:balance", bl)

	/*bl, err := stellar.GetAssetBalance(dest, "GRXT")
	log.Println("GetAssetBalance:receipt:balance", bl)
	*/

	// res1, res2, err := assets.SendAsset(asset, dest, float64(1000), issuerSeed, "init")
	// log.Println("SendAsset:", res1, res2, err)
	// bl, err = stellar.GetAssetBalance(dest, "GRXT")
	// if err != nil {
	// 	log.Println("GetAssetBalance:err", err)
	// } else {
	// 	log.Println("GetAssetBalance:receipt:balance:1 ", bl)
	// }

}

// func TestTrustline(t *testing.T) {
// 	issuer, err := ms.CreateKeyPair()
// 	if err != nil {
// 		log.Println(err)
// 		t.Fail()
// 	}

// 	mary, err := ms.CreateKeyPair()
// 	if err != nil {
// 		log.Println(err)
// 		t.Fail()
// 	}

// 	log.Println("issuer address:", issuer.Address)
// 	log.Println("issuer seed:", issuer.Seed)

// 	log.Println("mary address:", mary.Address)
// 	log.Println("mary seed:", mary.Seed)

// 	// Create a custom asset with the code "USD" issued by some trusted issuer.
// 	GRXT := microstellar.NewAsset("GRXT", issuer.Address, microstellar.Credit4Type)

// 	// Create a trust line from an account to the asset, with a limit of 10000.
// 	err = ms.CreateTrustLine(issuer.Seed, GRXT, "")
// 	if err != nil {
// 		log.Println("CreateTrustLine err:", err)
// 		t.Fail()
// 		return
// 	}
// 	// Make a payment in the asset.
// 	opt := new(microstellar.Options)
// 	err = ms.Pay(issuer.Seed, mary.Address, "10", GRXT, opt.WithMemoText("funny payment"))
// 	if err != nil {
// 		log.Println("Pay err:", err)
// 		t.Fail()
// 		return
// 	}

// 	// Require trustlines to be authorized buy issuer.
// 	ms.SetFlags(issuer.Seed, microstellar.FlagsNone)

// 	// Authorize a trustline after it's been created
// 	err = ms.AllowTrust(issuer.Seed, mary.Address, "GRXT", true)
// 	if err != nil {
// 		log.Println("AllowTrust err:", err)
// 		t.Fail()
// 	}
// }
