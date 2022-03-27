package api

const (
	DEFAULT_LIMIT  = 100
	DEFAULT_OFFSET = 0
)

// Define all error codes send back to clients
// This helps debug easier
const (
	SUPER_ADMIN_EMAIL = "grayll@grayll.io"
	//SUPER_ADMIN_EMAIL = "huykbc@gmail.com"
	SUPER_ADMIN_NAME = "GRAYLL"
)
const (
	UNKNOWN_ERROR          = "UNKNOWN_ERROR"
	TOKEN_EXPIRED          = "TOKEN_EXPIRED"
	INTERNAL_ERROR         = "INTERNAL_ERROR"
	NO_TOKEN               = "NO_TOKEN"
	TOKEN_INVALID          = "TOKEN_INVALID"
	INVALID_PARAMS         = "INVALID_PARAMS"
	DUPLICATE_NAME         = "DUPLICATE_NAME"
	INVALID_CODE           = "INVALID_CODE"
	EMAIL_IN_USED          = "EMAIL_IN_USED"
	EMAIL_INVALID          = "EMAIL_INVALID"
	INVALID_UNAME_PASSWORD = "INVALID_UNAME_PASSWORD"
	UNVERIFIED             = "UNVERIFIED"
	IP_CONFIRM             = "IP_CONFIRM"
	EMAIL_NOT_EXIST        = "EMAIL_NOT_EXIST"
	EMAIL_VERIFIED         = "EMAIL_VERIFIED"
	ADDRESS_EXIST          = "ADDRESS_EXIST"
	INVALID_ADDRESS        = "INVALID_ADDRESS"
	SUCCESS                = "SUCCESS"
	PHONE_EXIST            = "PHONE_EXIST"
	TX_FAIL                = "TX_FAIL"
	PRICE_LOWER_LIMIT      = "PRICE_LOWER_LIMIT"
	INTERNAL_ERROR_DES     = "INTERNAL_ERROR_DES"
	UNAPPROVED_EXIST       = "UNAPPROVED_EXIST"
	APPROVED               = "APPROVED"
	NOT_FOUND              = "NOT_FOUND"
)
const (
	GovPassport       = "GovPassport"
	GovNationalIdCard = "GovNationalIdCard"
	GovDriverLicense  = "GovDriverLicense"

	// requires tax return
	Income6MPaySlips   = "Income6MPaySlips"
	Income6MBankStt    = "Income6MBankStt"
	Income2YTaxReturns = "Income2YTaxReturns"

	// requires at least 2 docs
	AddressUtilityBill        = "AddressUtilityBill"
	AddressBankStt            = "AddressBankStt"
	AddressRentalAgreement    = "AddressRentalAgreement"
	AddressPropertyTaxReceipt = "AddressPropertyTaxReceipt"
	AddressTaxReturn          = "AddressTaxReturn"

	AssetsShareStockCert = "AssetsShareStockCert"
	Assets2MBankAccStt   = "Assets2MBankAccStt"
	Assets2MRetireAccStt = "Assets2MRetireAccStt"
	Assets2MInvestAccStt = "Assets2MInvestAccStt"

	// company documents
	CertIncorporation = "CertIncorporation"
	// require all docs
	Company2YTaxReturns       = "Company2YTaxReturns"
	Company2YFinancialStt     = "Company2YFinancialStt"
	Company2YBalanceSheets    = "Company2YBalanceSheets"
	Company6MBankStt          = "Company6MBankStt"
	Company6MInvestmentAccStt = "Company6MInvestmentAccStt"

	//Result audit

	GovPassportRes       = "GovPassportRes"
	GovNationalIdCardRes = "GovNationalIdCardRes"
	GovDriverLicenseRes  = "GovDriverLicenseRes"

	// requires tax return
	Income6MPaySlipsRes   = "Income6MPaySlipsRes"
	Income6MBankSttRes    = "Income6MBankSttRes"
	Income2YTaxReturnsRes = "Income2YTaxReturnsRes"

	// requires at least 2 docs
	AddressUtilityBillRes        = "AddressUtilityBillRes"
	AddressBankSttRes            = "AddressBankSttRes"
	AddressRentalAgreementRes    = "AddressRentalAgreementRes"
	AddressPropertyTaxReceiptRes = "AddressPropertyTaxReceiptRes"
	AddressTaxReturnRes          = "AddressTaxReturnRes"

	AssetsShareStockCertRes = "AssetsShareStockCertRes"
	Assets2MBankAccSttRes   = "Assets2MBankAccSttRes"
	Assets2MRetireAccSttRes = "Assets2MRetireAccSttRes"
	Assets2MInvestAccSttRes = "Assets2MInvestAccSttRes"

	// company documents
	CertIncorporationRes = "CertIncorporationRes"
	// require all docs
	Company2YTaxReturnsRes       = "Company2YTaxReturnsRes"
	Company2YFinancialSttRes     = "Company2YFinancialSttRes"
	Company2YBalanceSheetsRes    = "Company2YBalanceSheetsRes"
	Company6MBankSttRes          = "Company6MBankSttRes"
	Company6MInvestmentAccSttRes = "Company6MInvestmentAccSttRes"
)
const (
	API_KEY = "redacted"
	API_URL = "https://www.googleapis.com/identitytoolkit/v3/relyingparty/%s?key=redacted"
)
