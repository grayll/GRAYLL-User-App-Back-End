package mail

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	//"io/ioutil"
	"log"

	//"log"
	"os"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

const (
	grayllInfo     = "info@grayll.io"
	grayllInfoName = "GRAYLL Info"
	v3             = "/v3/mail/send"
	sendgridHost   = "https://api.sendgrid.com"
)

type MailInfo struct {
	to      string
	name    string
	subject string
	mode    string
	code    string
	mores   map[string]string
}

func BuildMail(to, name, subject, contentHtml string) []byte {
	toMail := mail.NewEmail(name, to)
	from := mail.NewEmail(grayllInfoName, grayllInfo)
	content := mail.NewContent("text/html", contentHtml)
	m := mail.NewV3MailInit(from, subject, toMail, content)

	spamCheckSetting := mail.NewSpamCheckSetting()
	spamCheckSetting.SetEnable(true)
	spamCheckSetting.SetSpamThreshold(1)
	spamCheckSetting.SetPostToURL("https://spamcatcher.sendgrid.com")
	mailSettings := mail.NewMailSettings()
	mailSettings.SetSpamCheckSettings(spamCheckSetting)
	m.SetMailSettings(mailSettings)

	return mail.GetRequestBody(m)
}

func SendMail(to, name, subject, mode, code, host string, mores map[string]string) error {
	api_key := os.Getenv("SENDGRID_API_KEY")
	request := sendgrid.GetRequest(api_key, v3, sendgridHost)
	request.Method = "POST"
	mailSerivce := NewEmailSerivce()
	htmlContent := ""
	var err error
	url := fmt.Sprintf("%s/handle?mode=%s&oobCode=%s", host, mode, code)
	switch mode {
	case "verifyEmail":
		log.Println("verifyEmail")
		htmlContent, err = mailSerivce.GenConfirmRegistration(name, url, 3)
	case "resetPassword":
		htmlContent, err = mailSerivce.GenResetPassword(name, url, 3)
	case "confirmIp":
		htmlContent, err = mailSerivce.GenConfirmIp(name, url, 3, mores)
	case "revealSecretToken":
		htmlContent, err = mailSerivce.GenRandomNumber(name, code)
	case "changeEmail":
		htmlContent, err = mailSerivce.GenChangeEmail(name, url, 3, mores["newemail"])
	default:
		log.Println("Not go any case email")
		return errors.New("Not found any email case")
	}

	if err != nil {
		return err
	}

	body := BuildMail(to, name, subject, htmlContent)
	request.Body = body
	_, err = sendgrid.API(request)
	if err != nil {
		fmt.Println("sendgrid.API error: ", err)
		return err
	}
	return nil
}

type Receipts struct {
	ReceiptsId []string `json:"persisted_recipients"`
}

func SendNoticeMail(to, name, subject string, contents []string) error {
	api_key := os.Getenv("SENDGRID_API_KEY")
	request := sendgrid.GetRequest(api_key, v3, sendgridHost)
	request.Method = "POST"
	mailSerivce := NewEmailSerivce()
	htmlContent, err := mailSerivce.GenMailWithContents(name, contents)
	if err != nil {
		return err
	}

	body := BuildMail(to, name, subject, htmlContent)
	request.Body = body
	_, err = sendgrid.API(request)
	if err != nil {
		fmt.Println("sendgrid.API error: ", err)
		return err
	}
	return nil
}
func SendMailRegistrationInvite(to, name, subject, url string, contents []string) error {
	api_key := os.Getenv("SENDGRID_API_KEY")
	request := sendgrid.GetRequest(api_key, v3, sendgridHost)
	request.Method = "POST"
	mailSerivce := NewEmailSerivce()
	htmlContent, err := mailSerivce.GenMailWithContentsAction(name, url, contents)
	if err != nil {
		return err
	}

	body := BuildMail(to, name, subject, htmlContent)
	request.Body = body
	_, err = sendgrid.API(request)
	if err != nil {
		fmt.Println("sendgrid.API error: ", err)
		return err
	}
	return nil
}
func SaveRegistrationInfo(name, lname, email string, createdTime, listId int64) (string, error) {
	apiKey := os.Getenv("SENDGRID_apiKey")
	request := sendgrid.GetRequest(apiKey, "/v3/contactdb/recipients", sendgridHost)
	request.Method = "POST"
	ts := time.Unix(createdTime, 0).Format("02-01-2006 15:04")
	jsonStr := fmt.Sprintf(`[{"last_name": "%s", "first_name": "%s", "email": "%s", "acc_created_at": "%s"}]`, lname, name, email, ts)
	recipientId := ""
	request.Body = []byte(jsonStr)
	res, err := sendgrid.API(request)
	if err != nil {
		fmt.Println("sendgrid.API error: ", err)
		fmt.Println(res.StatusCode)
		return "", err
	}

	receipts := Receipts{}
	err = json.Unmarshal([]byte(res.Body), &receipts)
	if err != nil {
		log.Println("Parse Receipts error: ", err)
		return "", err
	}
	if len(receipts.ReceiptsId) == 0 {
		log.Println("ERROR SaveRegistrationInfo - sendgrid returns empty receiptId", email, ts)
		return "", errors.New("sendgrid return empty receiptId")
	}
	recipientId = receipts.ReceiptsId[0]
	if listId > 0 {
		err = AddRecipienttoList(recipientId, listId)
		if err != nil {
			log.Println("AddRecipienttoList error: ", err)
			return "", err
		}
	} else {
		err = AddRecipienttoList(recipientId, 10196670)
		if err != nil {
			log.Println("AddRecipienttoList error: ", err)
			return "", err
		}
	}

	return recipientId, nil
}

// First Name
// Last Name
// Email Address
// Date App Account Creation (dd/mm/yyyy + hh:mm)
// Email Reminders Sent (#)
// Loan Paid (Yes/No)
func SaveLoanPaidInfo(name, lname, email, loanPaid string, createdTime, orderId int64) error {
	apiKey := os.Getenv("SENDGRID_API_KEY")
	ts := time.Unix(createdTime, 0).Format("02-01-2006 15:04")
	if orderId == 1 {
		request := sendgrid.GetRequest(apiKey, "/v3/contactdb/recipients", sendgridHost)
		request.Method = "POST"
		jsonStr := fmt.Sprintf(`[{"last_name": "%s", "first_name": "%s", "email": "%s", "acc_created_at": "%s", "loan_paid":"%s", "email_sents":%d}]`,
			lname, name, email, ts, loanPaid, orderId)

		request.Body = []byte(jsonStr)
		res, err := sendgrid.API(request)
		if err != nil {
			fmt.Println("sendgrid.API error: ", err)
			fmt.Println(res.StatusCode)
			return err
		} else {
			fmt.Println(res.Body)
			receipts := Receipts{}
			err = json.Unmarshal([]byte(res.Body), &receipts)
			if err != nil {
				log.Println("Parse Receipts error: ", err)
				return err
			}

			log.Println("receipts.ReceiptsId[0]", receipts.ReceiptsId[0])

			err = AddRecipienttoList(receipts.ReceiptsId[0], 10761027)
			if err != nil {
				log.Println("AddRecipienttoList error: ", err)
				return err
			}
		}
	} else {
		request := sendgrid.GetRequest(apiKey, "/v3/contactdb/recipients", sendgridHost)
		request.Method = "PATCH"
		jsonStr := fmt.Sprintf(`[{"email": "%s", "loan_paid": "%s", "email_sents":%d}]`, email, loanPaid, orderId)
		request.Body = []byte(jsonStr)
		res, err := sendgrid.API(request)
		if err != nil {
			fmt.Println("sendgrid.API error: ", err)
			fmt.Println(res.StatusCode)
			return err
		}
	}
	return nil
}

// AddaSingleRecipienttoaList : Add a Single Recipient to a List
// POST /contactdb/lists/{list_id}/recipients/{recipient_id}
// xlm loan 10761027
// grayll user 10196670
func AddRecipienttoList(recipient_id string, list_id int64) error {
	apiKey := os.Getenv("SENDGRID_API_KEY")
	request := sendgrid.GetRequest(apiKey, fmt.Sprintf("/v3/contactdb/lists/%d/recipients/%s", list_id, recipient_id), sendgridHost)
	request.Method = "POST"
	_, err := sendgrid.API(request)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
func RemoveRecipientFromList(recipient_id string, list_id int64) error {
	apiKey := os.Getenv("SENDGRID_API_KEY")
	request := sendgrid.GetRequest(apiKey, fmt.Sprintf("/v3/contactdb/lists/%d/recipients/%s", list_id, recipient_id), sendgridHost)
	request.Method = "DELETE"
	_, err := sendgrid.API(request)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func CreateCustomField() error {
	apiKey := os.Getenv("SENDGRID_API_KEY")
	request := sendgrid.GetRequest(apiKey, "/v3/contactdb/custom_fields", sendgridHost)
	request.Method = "POST"

	jsonStr := `{"name": "loan_paid", "type": "text"}`
	request.Body = []byte(jsonStr)
	res, err := sendgrid.API(request)
	if err != nil {
		fmt.Println("sendgrid.API custom_fields error: ", err)
		fmt.Println(res.StatusCode)
		return err
	}

	request = sendgrid.GetRequest(apiKey, "/v3/contactdb/custom_fields", sendgridHost)
	request.Method = "POST"

	jsonStr = `{"name": "acc_created_at", "type": "text"}`
	request.Body = []byte(jsonStr)
	res, err = sendgrid.API(request)
	if err != nil {
		fmt.Println("sendgrid.API custom_fields error: ", err)
		fmt.Println(res.StatusCode)
		return err
	}

	request = sendgrid.GetRequest(apiKey, "/v3/contactdb/custom_fields", sendgridHost)
	request.Method = "POST"

	jsonStr = `{"name": "email_sents", "type": "number"}`
	request.Body = []byte(jsonStr)
	res, err = sendgrid.API(request)
	if err != nil {
		fmt.Println("sendgrid.API custom_fields error: ", err)
		fmt.Println(res.StatusCode)
		return err
	}
	return nil
}

// func SendNoticeMail(to, name, subject string, contents []string) error {
// 	api_key := os.Getenv("SENDGRID_API_KEY")
// 	request := sendgrid.GetRequest(api_key, v3, sendgridHost)
// 	request.Method = "POST"
// 	mailSerivce := NewEmailSerivce()
// 	htmlContent, err := mailSerivce.GenMailWithContents(name, contents)
// 	if err != nil {
// 		return err
// 	}

// 	body := BuildMail(to, name, subject, htmlContent)
// 	request.Body = body
// 	_, err = sendgrid.API(request)
// 	if err != nil {
// 		fmt.Println("sendgrid.API error: ", err)
// 		return err
// 	}
// 	return nil
// }
func SendLoanReminder(to, name, subject, host string, contents []string, isPayoffButton bool) error {
	api_key := os.Getenv("SENDGRID_API_KEY")
	request := sendgrid.GetRequest(api_key, v3, sendgridHost)
	request.Method = "POST"
	mailSerivce := NewEmailSerivce()
	url := ""
	if isPayoffButton {
		url = fmt.Sprintf("%s/wallet/overview", host)
	} else {
		url = fmt.Sprintf("%s/register", host)
	}
	htmlContent, err := mailSerivce.GenMailLoanReminder(name, url, contents, isPayoffButton)
	if err != nil {
		return err
	}

	body := BuildMail(to, name, subject, htmlContent)
	request.Body = body
	_, err = sendgrid.API(request)
	if err != nil {
		fmt.Println("sendgrid.API error: ", err)
		return err
	}
	return nil
}
func SendMailResetPwdSuccess(to, name, subject string, contents []string) error {
	api_key := os.Getenv("SENDGRID_API_KEY")
	request := sendgrid.GetRequest(api_key, v3, sendgridHost)
	request.Method = "POST"
	mailSerivce := NewEmailSerivce()
	htmlContent, err := mailSerivce.GenMailWithContents(name, contents)
	if err != nil {
		return err
	}

	body := BuildMail(to, name, subject, htmlContent)
	request.Body = body
	_, err = sendgrid.API(request)
	if err != nil {
		fmt.Println("sendgrid.API error: ", err)
		return err
	}
	return nil
}

func SendMailLoanReminder(to, name, subject, url string, contents []string, isPayoffButton bool) error {
	api_key := os.Getenv("SENDGRID_API_KEY")
	request := sendgrid.GetRequest(api_key, v3, sendgridHost)
	request.Method = "POST"
	mailSerivce := NewEmailSerivce()
	htmlContent, err := mailSerivce.GenMailLoanReminder(name, url, contents, isPayoffButton)
	if err != nil {
		return err
	}

	body := BuildMail(to, name, subject, htmlContent)
	request.Body = body
	_, err = sendgrid.API(request)
	if err != nil {
		fmt.Println("sendgrid.API error: ", err)
		return err
	}
	return nil
}

func SendLoginNoticeMail(to, name, subject string, mores map[string]string) error {
	api_key := os.Getenv("SENDGRID_API_KEY")
	request := sendgrid.GetRequest(api_key, v3, sendgridHost)
	request.Method = "POST"
	mailSerivce := NewEmailSerivce()
	htmlContent, err := mailSerivce.GenLoginNotice(name, mores)
	if err != nil {
		return err
	}

	body := BuildMail(to, name, subject, htmlContent)
	request.Body = body
	_, err = sendgrid.API(request)
	if err != nil {
		fmt.Println("sendgrid.API error: ", err)
		return err
	}
	return nil
}
