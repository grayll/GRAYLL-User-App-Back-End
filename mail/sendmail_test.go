package mail

import (
	//"log"
	//"io/ioutil"
	//"log"
	"testing"
)

// func TestGenResetPassword(t *testing.T) {
// 	email := NewEmailSerivce()

// 	body, err := email.GenConfirmRegistration("Huy",
// 		"https://app.grayll.io/login/handle?mode=verifyEmail&oobCode=tAb8gxCCVdWLUWGBgARPIoOoj3D0Ds9kAwu1dJQYBKwAAAFs-yqyMQ&apiKey=AIzaSyCRH4tYsI1WdY652VF7Kpquu2_EYeC1yNc&lang=en",
// 		3)
// 	if err != nil {
// 		log.Println("GenResetPassword error: ", err)
// 		t.Fail()
// 	}
// 	log.Println("contentbody: ", body)

// 	err = ioutil.WriteFile("mail.html", []byte(body), 0664)
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// }

func TestSendMail(t *testing.T) {
	//email := NewEmailSerivce()

	// body, err := email.GenResetPassword("Huy",
	// 	"https://app.grayll.io/login/handle?mode=verifyEmail&oobCode=tAb8gxCCVdWLUWGBgARPIoOoj3D0Ds9kAwu1dJQYBKwAAAFs-yqyMQ&apiKey=AIzaSyCRH4tYsI1WdY652VF7Kpquu2_EYeC1yNc&lang=en",
	// 	3)
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// SendMail("huykbc@outlook.com", "Huy", "Confirm your registration with GRAYLL",
	// 	"verifyEmail", "tAb8gxCCVdWLUWGBgARPIoOoj3D0Ds9kAwu1dJQYBKwAAAF")

	SendNoticeMail("huykbc@gmail.com", "Huy", "Confirm your registration with GRAYLL", []string{"aa", "bb"})
}

func TestSaveRegistrationInfo(t *testing.T) {
	//CreateCustomField()
	//SaveRegistrationInfo("huy", "ngt", "huykbc@gmail.com", 1580313365)
	//SaveLoanPaidInfo("huy", "ngt", "huykbc@gmail.com", "yes", 1580313365, 2)
}
