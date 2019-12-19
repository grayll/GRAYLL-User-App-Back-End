package models

import (
	"time"

	"github.com/asaskevich/govalidator"
)

type (
	User struct {
		Id        int64     `json:"userId"`
		UserName  string    `json:"userName"`
		Password  string    `json:"password"`
		Email     string    `json:"email"`
		CreatedAt time.Time `json:"createdAt"`
	}
)

type Settings struct {
	IpConfirm    bool `json:"IpConfirm,omitempty"`
	MulSignature bool `json:"MulSignature,omitempty"`
	AppGeneral   bool `json:"AppGeneral,omitempty"`
	AppWallet    bool `json:"AppWallet,omitempty"`
	AppAlgo      bool `json:"AppAlgo,omitempty"`
	MailGeneral  bool `json:"MailGeneral,omitempty"`
	MailWallet   bool `json:"MailWallet,omitempty"`
	MailAlgo     bool `json:"MailAlgo,omitempty"`
}
type Tfa struct {
	BackupCode string `json:"BackupCode"`
	DataURL    string `json:"DataURL"`
	Enable     bool   `json:"Enable"`
	Secret     string `json:"Secret"`
	Expire     int64  `json:"Expire,omitempty"`
}
type TfaUpdate struct {
	BackupCode      string `json:"BackupCode"`
	DataURL         string `json:"DataURL"`
	Enable          bool   `json:"Enable"`
	Secret          string `json:"Secret"`
	OneTimePassword string `json:"OneTimePassword"`
}

type UserInfo struct {
	//Uid   string `json:"Uid,omitempty"`
	Name  string `json:"Name"`
	LName string `json:"LName,omitempty"`
	Email string `json:"Email"`
	//Token string `json:"Token,omitempty"`
	//Tfa                Tfa     `json:"Tfa"`
	// Number unread notice wallet,algo and general
	UrWallet       int      `json:"UrWallet"`
	UrAlgo         int      `json:"UrAlgo"`
	UrGeneral      int      `json:"UrGeneral"`
	HashPassword   string   `json:"HashPassword"`
	Setting        Settings `json:"Setting"`
	Ip             string   `json:"Ip,omitempty"`
	CreatedAt      int64    `json:"CreatedAt,omitempty"`
	PublicKey      string   `json:"PublicKey,omitempty"`
	LoanPaidStatus int      `json:"LoanPaidStatus,omitempty"`
	//SecretKeySalt      string   `json:"SecretKeySalt,omitempty""`
	Federation string `json:"Federation,omitempty"`
	IsVerified bool   `json:"IsVerified,omitempty"`
}
type UserLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (u *UserLogin) Validate() bool {
	if govalidator.IsNull(u.Password) || !govalidator.IsEmail(u.Email) {
		return false
	}
	return true
}

// Validate checks user data is valid or not for register.
func (u *UserInfo) Validate() bool {
	if govalidator.IsNull(u.Name) ||
		//govalidator.IsNull(u.PublicKey) || govalidator.IsNull(u.EncryptedSecretKey) ||
		!govalidator.IsEmail(u.Email) {
		return false
	}
	return true
}

// Validate checks user data is valid or not for login.
func (u *User) ValidateLogin() bool {
	if govalidator.IsNull(u.Password) || !govalidator.IsEmail(u.Email) {
		return false
	}
	return true
}
