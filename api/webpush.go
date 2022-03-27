package api

import (
	"encoding/json"

	webpush "github.com/SherClockHolmes/webpush-go"
)

func PushNotice(data map[string]interface{}, sub *webpush.Subscription) error {
	decoded, err := json.Marshal(data)
	_, err = webpush.SendNotification(decoded, sub, &webpush.Options{
		Subscriber:      "info@grayll.io",
		VAPIDPrivateKey: "redacted",
		VAPIDPublicKey:  "redacted",
		TTL:             30,
	})
	return err
}
