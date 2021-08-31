package api

import (
	"time"
	//"encoding"
	"encoding/base64"
	"fmt"
	//"time"
)

// func GenDocAccepted(name, lname, userId, accountId, appType, fieldName string, xlm, grx, algoValue float64) (string, string, []string) {
// 	docName := GetFriendlyName(fieldName)
// 	title := fmt.Sprintf(`GRAYLL | %s %s | %s | %s Submission Accepted`, name, lname, appType)
// 	contents := []string{
// 		time.Unix(time.Now().Unix(), 0).Format(`15:04 | 02-01-2006`),
// 		fmt.Sprintf(`%s %s has completed their %s  KYC submission for an administrator to review.`, name, lname, appType),

// 		`User Account: ` + accountId,

// 		`GRAYLL User ID: ` + userId,

// 		`====================`,

// 		fmt.Sprintf(`XLM Balance | $ %.4f`, xlm),
// 		fmt.Sprintf(`GRX Balance | $ %.4f`, grx),
// 		fmt.Sprintf(`USD Algo Position Value | $ %.4f`, algoValue),
// 	}

// 	content := ""
// 	for i, sent := range contents {
// 		content = content + sent
// 		if i == 0 {
// 			content = content + ". "
// 		}
// 	}

// 	return title, content, contents
// }

func GenDocAcceptedGrayll(name, lname, userId, appType, docName string, userValue UserValue) (string, string, []string) {
	title := fmt.Sprintf(`GRAYLL KYC | %s %s | %s KYC Document | %s Submission Accepted`, name, lname, appType, docName)
	contents := []string{
		time.Unix(time.Now().Unix(), 0).Format(`15:04 | 02-01-2006`),
		fmt.Sprintf(`A KYC Administrator has accepted the %s for a %s KYC submission of %s %s.`, docName, appType, name, lname),

		`User Account: ` + userValue.pk,

		`GRAYLL User ID: ` + userId,

		`====================`,

		fmt.Sprintf(`XLM Balance | XLM %.4f`, userValue.xlm),
		fmt.Sprintf(`GRX Balance | GRX %.4f`, userValue.grx),
		fmt.Sprintf(`GRY Balance | GRY %.4f`, userValue.gry),
		fmt.Sprintf(`USDC Balance | USDC %.4f`, userValue.usdc),
		fmt.Sprintf(`USD DeFi System Value | $ %.4f`, userValue.algoValue),
	}
	content := ""
	for i, sent := range contents {
		content = content + sent
		if i == 0 {
			content = content + ". "
		}
	}

	return title, content, contents
}
func GenDocDeclined(appType, docName, deadline string) (string, string, []string) {

	title := fmt.Sprintf(`GRAYLL KYC | %s KYC Document | %s Submission Declined`, appType, docName)

	contents := []string{
		time.Unix(time.Now().Unix(), 0).Format(`15:04 | 02-01-2006`),
		fmt.Sprintf(`A KYC Administrator has declined the %s for your %s KYC submission.`, docName, appType),
		fmt.Sprintf(`Please resubmit a valid document as soon as possible to meet the KYC deadline of %s.`, deadline),

		`=============================================`,

		`GRAYLL is legally required for every existing and new client to complete a Know Your Customer form to comply with various Government regulations around the world. Many legal jurisdictions are increasing regulations which have some positive and negative aspects, this means we need to review personal and financial information to verify your identity and eligibility to use GRAYLL Services.`,

		`GRAYLL offers advanced and sophisticated financial technology services that some jurisdictions do not deem appropriate for persons or entities that are not considered accredited or certified investors.`,

		`If you do not complete the KYC process within 60 days, your account will eventually be closed. We will send you periodic notifications to remind you prior to any account closure.`,

		`All information you provide is for verification purposes only by appropriate persons and will be kept strictly private and confidential, you may review our general Privacy Policy here: https://grayll.io/privacy`,

		`If you choose to not complete the KYC process or have not submitted all required information by the deadline an auditor will net your GRAYLL App account value.`,

		`If after review of a fully completed KYC submission you are not considered an accredited or certified investor an auditor will net your GRAYLL App account value.`,

		`Netting your GRAYLL App account means that you will receive any difference of the USD value of GRX purchased within the GRAYLL App only (not on exchanges) and the USD value of GRX sold by you.`,

		`If the USD value of the GRX you have sold is greater than the USD of the GRX you have purchased you will not receive anything. No claims can be made on any profits made by the GRAYLL Intelligence or Algorithmic Services.`,

		`The sooner you complete KYC the sooner we can start the audit process, however it is very important to note that before we are legally able to transfer any value in USDC or XLM to you, 	we must at least have received a Government Issued ID and a Proof of Address via the KYC application section in the GRAYLL App. If you do not want to submit any KYC information you may continue to trade GRX on Stellar DEX exchanges.`,

		`Please note that the length and difficulty of an audit process depends on the number of transfers and transactions within the GRAYLL App account and any other linked accounts someone may have used to buy or sell GRX.`,

		`We anticipate that the audit process of all GRAYLL App users will take 4 to 6 months in total. If you have any questions please consult our FAQs on GRAYLL Support https://support.grayll.io/en/ and if you cannot find the answers you are looking for there you may consult GRAYLL Client Support via the messenger.`,
	}

	content := ""
	for i, sent := range contents {
		content = content + sent
		if i == 0 {
			content = content + ". "
		}
	}

	return title, content, contents
}
func GenDocDeclinedGrayll(name, lname, userId, appType, docName string, userValue UserValue) (string, string, []string) {

	title := fmt.Sprintf(`GRAYLL KYC | %s %s | %s KYC Document | %s Submission Declined`, name, lname, appType, docName)
	contents := []string{
		time.Unix(time.Now().Unix(), 0).Format(`15:04 | 02-01-2006`),
		fmt.Sprintf(`A KYC Administrator has declined the %s for a %s KYC submission of %s %s.`, docName, appType, name, lname),

		`User Account: ` + userValue.pk,

		`GRAYLL User ID: ` + userId,

		`====================`,

		fmt.Sprintf(`XLM Balance | XLM %.4f`, userValue.xlm),
		fmt.Sprintf(`GRX Balance | GRX %.4f`, userValue.grx),
		fmt.Sprintf(`GRY Balance | GRY %.4f`, userValue.gry),
		fmt.Sprintf(`USDC Balance | USDC %.4f`, userValue.usdc),
		fmt.Sprintf(`USD DeFi System Value | $ %.4f`, userValue.algoValue),
	}
	content := ""
	for i, sent := range contents {
		content = content + sent
		if i == 0 {
			content = content + ". "
		}
	}

	return title, content, contents
}

func GenSubmitCompletedGrayll(name, lname, userId, appType string, userValue UserValue) (string, string, []string) {

	title := fmt.Sprintf(`GRAYLL | %s %s | %s KYC Submission Completed for Admin Review `, name, lname, appType)
	contents := []string{
		time.Unix(time.Now().Unix(), 0).Format(`15:04 | 02-01-2006`),
		fmt.Sprintf(`%s %s has completed their %s  KYC submission for an administrator to review.`, name, lname, appType),

		`User Account: ` + userValue.pk,

		`GRAYLL User ID: ` + userId,

		`====================`,

		fmt.Sprintf(`XLM Balance | XLM %.4f`, userValue.xlm),
		fmt.Sprintf(`GRX Balance | GRX %.4f`, userValue.grx),
		fmt.Sprintf(`GRY Balance | GRY %.4f`, userValue.gry),
		fmt.Sprintf(`USDC Balance | USDC %.4f`, userValue.usdc),
		fmt.Sprintf(`USD DeFi System Value | $ %.4f`, userValue.algoValue),
	}

	content := ""
	for i, sent := range contents {
		content = content + sent
		if i == 0 {
			content = content + ". "
		}
	}

	return title, content, contents
}
func GenSubmitCompleted(appType string) (string, string, []string) {
	title := fmt.Sprintf(`%s KYC Document Submission Completed`, appType)

	contents := []string{
		time.Unix(time.Now().Unix(), 0).Format(`15:04 | 02-01-2006`),
		fmt.Sprintf(`You have submitted all required documents for your %s KYC review.`, appType),

		`An administrator will review your documents within approximately 14 days, if any documents are declined or invalid you will receive notifications to re-submit those documents.`,

		`All information you provide is for verification purposes only by appropriate persons and will be kept strictly private and confidential, you may review our general Privacy Policy here: https://grayll.io/privacy`,

		`If after review of a fully completed KYC submission you are not considered an accredited or certified investor an auditor will net your GRAYLL App account value.`,

		`Netting your GRAYLL App account means that you will receive any difference of the USD value of GRX purchased within the GRAYLL App only (not on exchanges) and the USD value of GRX sold by you.`,

		`If the USD value of the GRX you have sold is greater than the USD of the GRX you have purchased you will not receive anything. No claims can be made on any profits made by the GRAYLL Intelligence or Algorithmic Services.`,

		`It is very important to note that before we are legally able to transfer any value in USDC or XLM to you, we must at least have received a Government Issued ID and a Proof of Address via the KYC application section in the GRAYLL App.`,

		`If you do not want to submit any KYC information you may continue to trade GRX on Stellar DEX exchanges.`,

		`Please note that the length and difficulty of an audit process depends on the number of transfers and transactions within the GRAYLL App account and any other linked accounts someone may have used to buy or sell GRX.`,

		`We anticipate that the audit process of all GRAYLL App users will take 4 to 6 months in total. If you have any questions please consult our FAQs on GRAYLL Support https://support.grayll.io/en/ and if you cannot find the answers you are looking for there you may consult GRAYLL Client Support via the messenger.`,
	}
	content := ""
	for i, sent := range contents {
		content = content + sent
		if i == 0 {
			content = content + ". "
		}
	}

	return title, content, contents
}

func GenDocSubmitOk(appType, docName string, lackDocs []string) (string, string, []string) {

	title := fmt.Sprintf(`%s KYC Document Submission Successful`, appType)

	contents := []string{
		time.Unix(time.Now().Unix(), 0).Format(`15:04 | 02-01-2006`),
		fmt.Sprintf(`Your %s document submission for your %s KYC review was successful.`, docName, appType),

		`You will also still need to submit the following documents to complete your GRAYLL KYC review:`,
	}

	contents1 := []string{

		`=============================================`,

		`GRAYLL is legally required for every existing and new client to complete a Know Your Customer form to comply with various Government regulations around the world. Many legal jurisdictions are increasing regulations which have some positive and negative aspects, this means we need to review personal and financial information to verify your identity and eligibility to use GRAYLL Services.`,

		`GRAYLL offers advanced and sophisticated financial technology services that some jurisdictions do not deem appropriate for persons or entities that are not considered accredited or certified investors.`,

		`If you do not complete the KYC process within 60 days, your account will eventually be closed. We will send you periodic notifications to remind you prior to any account closure.`,

		`All information you provide is for verification purposes only by appropriate persons and will be kept strictly private and confidential, you may review our general Privacy Policy here: https://grayll.io/privacy`,

		`If you choose to not complete the KYC process or have not submitted all required information by the deadline an auditor will net your GRAYLL App account value.`,

		`If after review of a fully completed KYC submission you are not considered an accredited or certified investor an auditor will net your GRAYLL App account value.`,

		`Netting your GRAYLL App account means that you will receive any difference of the USD value of GRX purchased within the GRAYLL App only (not on exchanges) and the USD value of GRX sold by you.`,

		`If the USD value of the GRX you have sold is greater than the USD of the GRX you have purchased you will not receive anything. No claims can be made on any profits made by the GRAYLL Intelligence or Algorithmic Services.`,

		`The sooner you complete KYC the sooner we can start the audit process, however it is very important to note that before we are legally able to transfer any value in USDC or XLM to you, 	we must at least have received a Government Issued ID and a Proof of Address via the KYC application section in the GRAYLL App. If you do not want to submit any KYC information you may continue to trade GRX on Stellar DEX exchanges.`,

		`Please note that the length and difficulty of an audit process depends on the number of transfers and transactions within the GRAYLL App account and any other linked accounts someone may have used to buy or sell GRX.`,

		`We anticipate that the audit process of all GRAYLL App users will take 4 to 6 months in total. If you have any questions please consult our FAQs on GRAYLL Support https://support.grayll.io/en/ and if you cannot find the answers you are looking for there you may consult GRAYLL Client Support via the messenger.`,
	}

	contents = append(contents, lackDocs...)
	contents = append(contents, contents1...)

	content := ""
	for i, sent := range contents {
		content = content + sent
		if i == 0 {
			content = content + ". "
		}
	}

	return title, content, contents
}

func GenFinalDeclinedGrayll(name, lname, userId, appType string, userValue UserValue) (string, string, []string) {

	title := fmt.Sprintf(`GRAYLL | KYC %s Application | %s %s | Declined`, appType, name, lname)
	contents := []string{
		time.Unix(time.Now().Unix(), 0).Format(`15:04 | 02-01-2006`),
		fmt.Sprintf(`A KYC Administrator has declined the %s KYC Application of %s %s.`, appType, name, lname),

		`User Account: ` + userValue.pk,

		`GRAYLL User ID: ` + userId,

		`====================`,

		fmt.Sprintf(`XLM Balance | XLM %.4f`, userValue.xlm),
		fmt.Sprintf(`GRX Balance | GRX %.4f`, userValue.grx),
		fmt.Sprintf(`GRY Balance | GRY %.4f`, userValue.gry),
		fmt.Sprintf(`USDC Balance | USDC %.4f`, userValue.usdc),
		fmt.Sprintf(`USD DeFi System Value | $ %.4f`, userValue.algoValue),
	}
	content := ""
	for i, sent := range contents {
		content = content + sent
		if i == 0 {
			content = content + ". "
		}
	}

	return title, content, contents
}

func GenFinalApproveGrayll(name, lname, userId, appType string, userValue UserValue) (string, string, []string) {

	title := fmt.Sprintf(`GRAYLL | KYC %s Application | %s %s | Approved`, appType, name, lname)
	contents := []string{
		time.Unix(time.Now().Unix(), 0).Format(`15:04 | 02-01-2006`),
		fmt.Sprintf(`A KYC Administrator has approved the %s KYC Application of %s %s.`, appType, name, lname),

		`User Account: ` + userValue.pk,

		`GRAYLL User ID: ` + userId,

		`====================`,

		fmt.Sprintf(`XLM Balance | XLM %.4f`, userValue.xlm),
		fmt.Sprintf(`GRX Balance | GRX %.4f`, userValue.grx),
		fmt.Sprintf(`GRY Balance | GRY %.4f`, userValue.gry),
		fmt.Sprintf(`USDC Balance | USDC %.4f`, userValue.usdc),
		fmt.Sprintf(`USD DeFi System Value | $ %.4f`, userValue.algoValue),
	}
	content := ""
	for i, sent := range contents {
		content = content + sent
		if i == 0 {
			content = content + ". "
		}
	}

	return title, content, contents
}

func GenFinalDeclined(name, lname, userId, accountId, appType string) (string, string, []string) {

	title := fmt.Sprintf(`GRAYLL | KYC %s Application | %s %s | Declined`, appType, name, lname)
	contents := []string{
		time.Unix(time.Now().Unix(), 0).Format(`15:04 | 02-01-2006`),
		fmt.Sprintf(`A KYC Administrator has declined the %s KYC Application completed by %s %s.`, appType, name, lname),

		`User Account: ` + accountId,

		`GRAYLL User ID: ` + userId,

		`  `,

		`After careful review and consideration you do not qualify as an accredited investor within the definition set by the Canadian authorities;`,

		`1) An individual, alone or with a spouse, who has net assets of more than $5 million CAD.`,

		`2) A person registered in Canada, under securities legislation, as a dealer or an adviser.`,

		`An auditor will net your GRAYLL App account value, netting your GRAYLL App account means that you will receive any difference of the USD value of GRX purchased within the GRAYLL App only (not on exchanges) and the USD value of GRX sold by you.`,

		`If the USD value of the GRX you have sold is greater than the USD of the GRX you have purchased you will not receive anything. No claims can be made on any profits made by the GRAYLL Intelligence or Algorithmic Services.`,

		`It is very important to note that before we are legally able to transfer any value in USDC or XLM to you, we must at least have received a Government Issued ID and a Proof of Address via the KYC application section in the GRAYLL App. `,

		`If you do not want to submit any KYC information you may continue to trade GRX on Stellar DEX exchanges.`,

		`Please note that the length and difficulty of an audit process depends on the number of transfers and transactions within the GRAYLL App account and any other linked accounts someone may have used to buy or sell GRX.`,

		`We anticipate that the audit process of all GRAYLL App users will take 4 to 6 months in total. If you have any questions please consult our FAQs on GRAYLL Support https://support.grayll.io/en/ and if you cannot find the answers you are looking for there you may consult GRAYLL Client Support via the messenger.`,

		`Thank you for your time and patience, we hope that in the future we will also provide GRAYLL Services for non-accredited investors.`,
	}
	content := ""
	for i, sent := range contents {
		content = content + sent
		if i == 0 {
			content = content + ". "
		}
	}

	return title, content, contents
}

func GenFinalApprove(name, lname, userId, accountId, appType string) (string, string, []string) {

	title := fmt.Sprintf(`GRAYLL | KYC %s Application | %s %s | Approved`, appType, name, lname)
	contents := []string{
		time.Unix(time.Now().Unix(), 0).Format(`15:04 | 02-01-2006`),
		fmt.Sprintf(`A KYC Administrator has approved the %s KYC Application completed by %s %s.`, appType, name, lname),

		`User Account: ` + accountId,

		`GRAYLL User ID: ` + userId,
	}
	content := ""
	for i, sent := range contents {
		content = content + sent
		if i == 0 {
			content = content + ". "
		}
	}

	return title, content, contents
}

func GenKycRevokeGrayll(name, lname, userId, appType string, userValue UserValue) (string, string, []string) {

	title := fmt.Sprintf(`GRAYLL | KYC %s Application | %s %s | Revoke`, appType, name, lname)
	contents := []string{
		time.Unix(time.Now().Unix(), 0).Format(`15:04 | 02-01-2006`),
		fmt.Sprintf(`A KYC Administrator has revoked the %s KYC Application of %s %s.`, appType, name, lname),

		`User Account: ` + userValue.pk,

		`GRAYLL User ID: ` + userId,

		`====================`,

		fmt.Sprintf(`XLM Balance | XLM %.4f`, userValue.xlm),
		fmt.Sprintf(`GRX Balance | GRX %.4f`, userValue.grx),
		fmt.Sprintf(`GRY Balance | GRY %.4f`, userValue.gry),
		fmt.Sprintf(`USDC Balance | USDC %.4f`, userValue.usdc),
		fmt.Sprintf(`USD DeFi System Value | $ %.4f`, userValue.algoValue),
	}
	content := ""
	for i, sent := range contents {
		content = content + sent
		if i == 0 {
			content = content + ". "
		}
	}

	return title, content, contents
}

func GenKycRevoke(name, lname, userId, accountId, appType string) (string, string, []string) {

	title := fmt.Sprintf(`GRAYLL | KYC %s Application | %s %s | Revoked`, appType, name, lname)
	contents := []string{
		time.Unix(time.Now().Unix(), 0).Format(`15:04 | 02-01-2006`),
		fmt.Sprintf(`A KYC Administrator has revoked the %s KYC Approved Status for %s %s.`, appType, name, lname),

		`User Account: ` + accountId,

		`GRAYLL User ID: ` + userId,

		`You will be notified which documents and information you will need to submit to meet KYC requirements and continue to use GRAYLL Services.`,
	}
	content := ""
	for i, sent := range contents {
		content = content + sent
		if i == 0 {
			content = content + ". "
		}
	}

	return title, content, contents
}

//=========

func GenInvite(uid, name, lname, docId string) (string, string, []string) {

	encodedUid := base64.StdEncoding.EncodeToString([]byte(uid))

	url := fmt.Sprintf(`https://app.grayll.io/register?referer=%s&id=%s`, encodedUid, docId)
	title := fmt.Sprintf(`GRAYLL | Sign Up Invite — received from — %s %s`, name, lname)

	contents := []string{
		//fmt.Sprintf(`Dear %s %s`, name, lname),

		fmt.Sprintf(`%s %s has invited you to Sign Up for the GRAYLL App.`, name, lname),

		fmt.Sprintf(`GRAYLL is a Simple Automated Recession Proof Investment App — GRAYLL Intelligence advanced AI driven solutions automate Wealth Inception to help you increase your savings, investment & trading returns.`),

		fmt.Sprintf(`Time is the most valued & under-appreciated asset we have. GRAYLL is a Decentralized Digital System which automatically lets you profit from opportunities in a variety of markets - 24/7/365 - using the latest Artificial Intelligence technology advancements.`),

		fmt.Sprintf(`GRAYLL is an innovative Recession Proof & Anti-Fragile System. Investors, traders, savers and pensioners using GRAYLL Intelligence have a significant advantage in these catastrophic financial markets exacerbated by the COVID-19 Coronavirus Crisis.`),

		fmt.Sprintf(`By accepting this GRAYLL App Invite Request sent by your contact — you benefit from Discounted Performance Fees of 15%% on realized profits — the Standard Performance Fee on realized profits is 18%% — if you have any questions please do not hesitate to reach out via messenger on our website.`),
	}
	content := ""
	for i, sent := range contents {
		content = content + sent
		if i == 0 {
			content = content + ". "
		}
	}

	return title, url, contents
}

func GenInviteSender(name, lname string) (string, string, []string) {

	title := fmt.Sprintf(`GRAYLL | Sign Up Invite — sent to — %s %s`, name, lname)

	contents := []string{
		//fmt.Sprintf(`Dear %s %s`, name, lname),

		fmt.Sprintf(`We have sent your Referral Contact Invite request to %s %s.`, name, lname),

		fmt.Sprintf(`Until your Referral Contact has accepted your invite and signed up for a GRAYLL Account you will see this invite in Pending Invites within your GRAYLL App.`),

		fmt.Sprintf(`If your Referral Contact has not accepted your invite within a week, we recommend following up directly — before Sending a Reminder from the GRAYLL App — the invite may have simply landed in their spam folder.`),

		fmt.Sprintf(`A confirmed Referral Contact will benefit from a reduction in Performance Fees — instead of the Standard Performance Fee of 18%% — your Referral Contact will pay Discounted Performance Fees of 15%% on Algorithmic Position profits realized.`),

		fmt.Sprintf(`For every confirmed Referral Contact you will receive a share of the Performance Fees from the algorithmic positions that your confirmed Referral Contact has closed with realized profits.`),

		fmt.Sprintf(`You will receive 3%% of the realized profits and GRAYLL will receive 12%% of the realized profits.`),
	}
	content := ""
	for i, sent := range contents {
		content = content + sent
		if i == 0 {
			content = content + ". "
		}
	}

	return title, content, contents
}

func GenInvitationConfirmed(refername, referlname, inviteDate string) (string, string, []string) {

	title := fmt.Sprintf(`GRAYLL | Referral Invite Accepted — from — %s %s`, refername, referlname)

	contents := []string{

		fmt.Sprintf(`You have accepted the Referral Invite from %s %s that was sent on %s.`, refername, referlname, inviteDate),

		fmt.Sprintf(`You will benefit from a reduction in Performance Fees — instead of the Standard Performance Fee of 18%% — you will pay Discounted Performance Fees of 15%% on Algorithmic Position profits realized.`),
	}
	content := ""
	for i, sent := range contents {
		content = content + sent
		if i == 0 {
			content = content + ". "
		}
	}

	return title, content, contents
}

func GenInvitationConfirmedSender(name, lname, inviteDate string) (string, string, []string) {

	title := fmt.Sprintf(`GRAYLL | Referral Contact Invite Accepted — by — %s %s`, name, lname)

	contents := []string{

		fmt.Sprintf(`%s %s has accepted the Referral Invite you sent on %s.`, name, lname, inviteDate),

		fmt.Sprintf(`1) Your confirmed Referral Contact will benefit from a reduction in Performance Fees — instead of the Standard Performance Fee of 18%% — your Referral Contact will pay Discounted Performance Fees of 15%% on Algorithmic Position profits realized.`),

		fmt.Sprintf(`2) You will receive a share of the Performance Fees from the algorithmic positions that your confirmed Referral Contact has closed with realized profits.`),

		fmt.Sprintf(`3) You will receive 3%% of the realized profits and GRAYLL will receive 12%% of the realized profits.`),
	}
	content := ""
	for i, sent := range contents {
		content = content + sent
		if i == 0 {
			content = content + ". "
		}
	}

	return title, content, contents
}

func GenRemoveReferer(name, lname string) (string, string, []string) {

	title := fmt.Sprintf(`GRAYLL | Removed Linked Referrer Contact — %s %s`, name, lname)

	contents := []string{

		fmt.Sprintf(`You have removed %s %s as your Linked Referrer Contact from your GRAYLL App.`, name, lname),

		fmt.Sprintf(`Removing your Referrer Contact has resulted in the following:`),

		fmt.Sprintf(`1) You will no longer benefit from Discounted Performance Fees — your Performance Fee on realized profits has increased to 18%% — the Standard Performance Fee.`),

		fmt.Sprintf(`2) Your Referrer Contact will no longer receive a share of your Performance Fees from your algo positions closed with realized profits.`),

		fmt.Sprintf(`This action does not affect your other Referral Contacts — all your confirmed Referral Contacts will continue to receive a discounted Performance Fee of 15%% on realized profits.`),

		fmt.Sprintf(`This action also does not affect the Referral Fees you receive from your confirmed Referral Contacts when they close algo positions with realized profits.`),
	}
	content := ""
	for i, sent := range contents {
		content = content + sent
		if i == 0 {
			content = content + ". "
		}
	}

	return title, content, contents
}

func GenRemoveRefererSender(name, lname string) (string, string, []string) {

	title := fmt.Sprintf(`GRAYLL | %s %s Removed You as Linked Referrer Contact`, name, lname)

	contents := []string{

		fmt.Sprintf(`%s %s has removed you as their Linked Referrer Contact.`, name, lname),

		fmt.Sprintf(`Being removed as the Linked Referrer Contact has resulted in the following:`),

		fmt.Sprintf(`1) You will no longer receive Referral Fees from %s %s  — this action does not affect your other Referral Contacts.`, name, lname),

		fmt.Sprintf(`2) You will continue to receive a 3%% share on the realized profits from algorithmic positions closed by your confirmed Referral Contacts.`),
	}
	content := ""
	for i, sent := range contents {
		content = content + sent
		if i == 0 {
			content = content + ". "
		}
	}

	return title, content, contents
}

func GenRemoveRefererralSender(name, lname string) (string, string, []string) {

	title := fmt.Sprintf(`GRAYLL | %s %s Removed You as Linked Referrer Contact`, name, lname)

	contents := []string{

		fmt.Sprintf(`%s %s has removed you as their Linked Referrer Contact.`, name, lname),

		fmt.Sprintf(`Being removed as the Linked Referrer Contact has resulted in the following:`),

		fmt.Sprintf(`1) You will no longer receive Referral Fees from %s %s  — this action does not affect your other Referral Contacts.`, name, lname),

		fmt.Sprintf(`2) You will continue to receive a 3%% share on the realized profits from algorithmic positions closed by your confirmed Referral Contacts.`),
	}
	content := ""
	for i, sent := range contents {
		content = content + sent
		if i == 0 {
			content = content + ". "
		}
	}

	return title, content, contents
}

func GenRemoveRefererral(name, lname string) (string, string, []string) {

	title := fmt.Sprintf(`GRAYLL | %s %s — Removed You as a Referral Contact`, name, lname)

	contents := []string{

		fmt.Sprintf(`%s %s has removed you as their Referral Contact.`, name, lname),

		fmt.Sprintf(`Being removed as the Referral Contact has resulted in the following:`),

		fmt.Sprintf(`1) You will no longer benefit from Discounted Performance Fees — you will now pay a Standard Performance Fee of 18%% — on Algorithmic Position profits realized.`),

		fmt.Sprintf(`2) You will continue to receive a 3%% share on the realized profits from algorithmic positions closed by your confirmed Referral Contacts.`),
	}
	content := ""
	for i, sent := range contents {
		content = content + sent
		if i == 0 {
			content = content + ". "
		}
	}

	return title, content, contents
}
func GenReminder(uid, name, lname, docId string) (string, string, []string) {

	encodedUid := base64.StdEncoding.EncodeToString([]byte(uid))

	url := fmt.Sprintf(`https://app.grayll.io/register?referer=%s&id=%s`, encodedUid, docId)
	title := fmt.Sprintf(`GRAYLL | Sign Up Invite — received from — %s %s`, name, lname)

	contents := []string{
		fmt.Sprintf(`%s %s has sent you a Reminder to Sign Up (%s) for the GRAYLL App.`, name, lname, url),

		fmt.Sprintf(`GRAYLL is a Simple Automated Recession Proof Investment App — GRAYLL Intelligence advanced AI driven solutions automate Wealth Inception to help you increase your savings, investment & trading returns.`),

		fmt.Sprintf(`Time is the most valued & under-appreciated asset we have. GRAYLL is a Decentralized Digital System which automatically lets you profit from opportunities in a variety of markets - 24/7/365 - using the latest Artificial Intelligence technology advancements.`),

		fmt.Sprintf(`GRAYLL is an innovative Recession Proof & Anti-Fragile System. Investors, traders, savers and pensioners using GRAYLL Intelligence have a significant advantage in these catastrophic financial markets exacerbated by the COVID-19 Coronavirus Crisis.`),

		fmt.Sprintf(`By accepting this GRAYLL App Invite Request sent by your contact — you benefit from Discounted Performance Fees of 15%% on realized profits — the Standard Performance Fee on realized profits is 18%% — if you have any questions please do not hesitate to reach out via messenger on our website.`),
	}
	content := ""
	for i, sent := range contents {
		content = content + sent
		if i == 0 {
			content = content + ". "
		}
	}

	return title, content, contents
}
func GenReminderSender(name, lname string) (string, string, []string) {

	title := fmt.Sprintf(`GRAYLL | Sign Up Invite Reminder Sent — to — %s %s`, name, lname)

	contents := []string{

		fmt.Sprintf(`We have sent your Invite Referral Reminder to your Referral Contact %s %s.`, name, lname),

		fmt.Sprintf(`Until your Referral Contact has accepted your invite and signed up for a GRAYLL Account you will see this invite in Pending Invites within your GRAYLL App.`),

		fmt.Sprintf(`If your Referral Contact has not accepted your invite or contacted you within a week, we recommend following up directly — before Sending another Reminder from the GRAYLL App — the invite may have simply landed in their spam folder.`),

		fmt.Sprintf(`A confirmed Referral Contact will benefit from a reduction in Performance Fees — instead of the Standard Performance Fee of 18%% — your Referral Contact will pay Discounted Performance Fees of 15%% on Algorithmic Position profits realized.`),

		fmt.Sprintf(`For every confirmed Referral Contact you will receive a share of the Performance Fees from the algorithmic positions that your confirmed Referral Contact has closed with realized profits.`),

		fmt.Sprintf(`You will receive 3%% of the realized profits and GRAYLL will receive 12%% of the realized profits.`),
	}
	content := ""
	for i, sent := range contents {
		content = content + sent
		if i == 0 {
			content = content + ". "
		}
	}

	return title, content, contents
}
