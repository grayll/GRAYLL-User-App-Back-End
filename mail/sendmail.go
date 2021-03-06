package mail

import (
	"fmt"
	"log"

	//"log"
	//"fmt"
	"time"

	"github.com/huyntsgs/hermes"
)

type Email struct {
	branch hermes.Hermes
	email  hermes.Email
}

var SendGridAPIKey string

func NewEmailSerivce() *Email {
	h := hermes.Hermes{
		// Optional Theme
		// Theme: new(Default)
		Product: hermes.Product{
			// Appears in header & footer of e-mails
			Name:      "GRAYLL",
			Link:      "https://grayll.io",
			Copyright: `Copyright © 2020, GRAYLL. All Rights Reserved.`,
			// 			Terms of Service | Privacy Policy`,
			// Optional product logo
			//Logo: "http://www.duchess-france.org/wp-content/uploads/2016/01/gopher.png",
		},
	}
	return &Email{branch: h}
}

func (email *Email) GenConfirmRegistration(firstName, url string, expiredInDays int) (string, error) {
	timeExpire := ""
	if expiredInDays > 0 {
		duration := time.Duration(expiredInDays*24) * time.Hour
		expDate := time.Now().Add(duration)
		timeExpire = expDate.Format("15:04 on January 2, 2006")
	}
	emailBody := hermes.Email{
		Body: hermes.Body{
			Name:      firstName,
			Greeting:  "Dear",
			Signature: "With gratitude",
			Intros: []string{
				"Thanks for joining GRAYLL!",
			},
			Actions: []hermes.Action{
				{
					Instructions: "Please click here to verify your email and get started.",
					Button: hermes.Button{
						Color: "#40278C", // Optional action button color
						Text:  "VERIFY EMAIL",
						Link:  url,
					},
				},
			},
			Outros: []string{
				"Once the account is activated, you can sign in and add Stellar Lumens (XLM) and GRAYLL GRX to your account",
				"Please notice that account activation link expires at " + timeExpire,
				"If you didn’t create a GRAYLL account, you can ignore this email.",
			},
		},
	}

	return email.branch.GenerateHTML(emailBody)
}

func (email *Email) GenResetPassword(firstName, url string, expiredInDays int) (string, error) {
	timeExpire := ""
	if expiredInDays > 0 {
		duration := time.Duration(expiredInDays*24) * time.Hour
		expDate := time.Now().Add(duration)
		timeExpire = expDate.Format("15:04 on January 2, 2006")
	}
	emailBody := hermes.Email{
		Body: hermes.Body{
			Name: firstName,
			Intros: []string{
				"Thanks for joining GRAYLL!",
			},
			Actions: []hermes.Action{
				{
					Instructions: "You have requested to reset your password. Please click here to initiate your password reset.",
					Button: hermes.Button{
						Color: "#40278C", // Optional action button color
						Text:  "RESET PASSWORD",
						Link:  url,
					},
				},
			},
			Outros: []string{
				"Please notice that reset password link expires at " + timeExpire,
				"If you didn’t request to reset your GRAYLL account password, you may ignore this email.",
			},
			Signature: "With gratitude",
			Greeting:  "Hello",
		},
	}

	return email.branch.GenerateHTML(emailBody)
}

func (email *Email) GenRandomNumber(firstName, ranstr string) (string, error) {
	emailBody := hermes.Email{
		Body: hermes.Body{
			Name:      firstName,
			Greeting:  "Dear",
			Signature: "With gratitude",
			Outros: []string{
				`We have received a Reveal Secret Key request from your account.`,

				`For security reasons this request and code is valid for 10 minutes.`,

				fmt.Sprintf(`Please copy and paste the following code in the form in your GRAYLL App: %s`, ranstr),

				`If this wasn't you - please contact us immediately! 	support@grayll.io`,
			},
		},
	}
	//SATVZJU3QDKSRVLDONMNFGURA3BUT2FT7HXRPABJ7PFOKQCTTOQZONG4
	return email.branch.GenerateHTML(emailBody)
}

func (email *Email) GenMailWithContents(firstName string, contents []string) (string, error) {
	emailBody := hermes.Email{
		Body: hermes.Body{
			Name:      firstName,
			Greeting:  "Dear",
			Signature: "With gratitude",
			Outros:    contents,
		},
	}
	return email.branch.GenerateHTML(emailBody)
}

func (email *Email) GenMailWithContentsAction(firstName, url string, contents []string) (string, error) {

	emailBody := hermes.Email{
		Body: hermes.Body{
			Name:      firstName,
			Greeting:  "Dear",
			Signature: "With gratitude",
			Intros:    contents,
			Actions: []hermes.Action{
				{
					//Instructions: "Please click here to verify your email and get started.",
					Button: hermes.Button{
						Color: "#40278C", // Optional action button color
						Text:  "SIGN UP",
						Link:  url,
					},
				},
			},
		},
	}
	return email.branch.GenerateHTML(emailBody)
}
func (email *Email) GenMailLoanReminder(firstName, url string, contents []string, isPayoffButton bool) (string, error) {
	var emailBody hermes.Email
	if isPayoffButton {
		emailBody = hermes.Email{
			Body: hermes.Body{
				Name:      firstName,
				Greeting:  "Dear",
				Signature: "With gratitude",
				Intros:    contents,
				Actions: []hermes.Action{
					{
						Button: hermes.Button{
							Color: "#40278C",
							Text:  "PAY OFF LOAN",
							Link:  url,
						},
					},
				},
			},
		}
	} else {
		emailBody = hermes.Email{
			Body: hermes.Body{
				Name:      firstName,
				Greeting:  "Dear",
				Signature: "With gratitude",
				Intros:    contents,
				Actions: []hermes.Action{
					{
						Button: hermes.Button{
							Color: "#40278C",
							Text:  "SIGN UP",
							Link:  url,
						},
					},
				},
			},
		}
	}
	return email.branch.GenerateHTML(emailBody)
}

func (email *Email) GenConfirmIp(firstName, url string, expiredInDays int, mores map[string]string) (string, error) {
	// timeExpire := ""
	// if expiredInDays > 0 {
	// 	duration := time.Duration(expiredInDays*24) * time.Hour
	// 	expDate := time.Now().Add(duration)
	// 	timeExpire = expDate.Format("15:04 on January 2, 2006")
	// }
	log.Println("mores:", mores)
	emailBody := hermes.Email{
		Body: hermes.Body{
			Name:      firstName,
			Greeting:  "Dear",
			Signature: "With gratitude",
			// Intros: []string{
			// 	"Thanks for joining GRAYLL!",
			// },
			Actions: []hermes.Action{
				// {
				// 	Instructions: "We have detected a successful login attempt to your account from an unused IP address.",
				// },
				{
					Instructions: `Your IP address was not recognised by the system, it might be that you have not used it for a certain period of time.
					Before accessing your account, please confirm your IP address by clicking the button below.`,
					Button: hermes.Button{
						Color: "#40278C", // Optional action button color
						Text:  "CONFIRM IP",
						Link:  url,
					},
				},
			},
			Outros: []string{
				"The login details are below:",
				"Login Time: " + mores["loginTime"],
				"User Agent: " + mores["agent"],
				"IP Address: " + mores["ip"],
				"City: " + mores["city"],
				"Country: " + mores["country"],
				"If this wasn't you - please contact us immediately!",
				"support@grayll.io",
			},
		},
	}

	return email.branch.GenerateHTML(emailBody)
}
func (email *Email) GenLoginNotice(firstName string, mores map[string]string) (string, error) {
	emailBody := hermes.Email{
		Body: hermes.Body{
			Name:      firstName,
			Greeting:  "Dear",
			Signature: "With gratitude",
			Outros: []string{
				"We have detected a successful login attempt to your account.",
				"The login details are below:",
				"Login Time: " + mores["loginTime"],
				"User Agent: " + mores["agent"],
				"IP Address: " + mores["ip"],
				"City: " + mores["city"],
				"Country: " + mores["country"],
				"If this wasn't you - please contact us immediately!",
				"support@grayll.io",
			},
		},
	}
	return email.branch.GenerateHTML(emailBody)
}
func (email *Email) GenChangeEmail(firstName, url string, expiredInDays int, newemail string) (string, error) {
	timeExpire := ""
	if expiredInDays > 0 {
		duration := time.Duration(expiredInDays*24) * time.Hour
		expDate := time.Now().Add(duration)
		timeExpire = expDate.Format("15:04 on January 2, 2006")
	}
	emailBody := hermes.Email{
		Body: hermes.Body{
			Name:      firstName,
			Greeting:  "Dear",
			Signature: "With gratitude",
			Actions: []hermes.Action{
				{
					Instructions: "We have received change email request to new email " + newemail + ". Please click here to change email",
					Button: hermes.Button{
						Color: "#40278C", // Optional action button color
						Text:  "CHANGE EMAIL",
						Link:  url,
					},
				},
			},
			Outros: []string{
				"Once you click to confirm change email, new email is udpated and require you to confirm registration with the new emai.",
				"Please notice that change email link expires at " + timeExpire,
				"If this wasn't you - please contact us immediately!",
			},
		},
	}

	return email.branch.GenerateHTML(emailBody)
}
