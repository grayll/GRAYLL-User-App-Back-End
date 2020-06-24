package api

import (
	//"encoding"
	"encoding/base64"
	"fmt"
	//"time"
)

func GenInvite(uid, name, lname, docId string) (string, string, []string) {

	encodedUid := base64.StdEncoding.EncodeToString([]byte(uid))

	url := fmt.Sprintf(`https://app.grayll.io/register?referer=%s&id=%s`, encodedUid, docId)
	title := fmt.Sprintf(`GRAYLL | Sign Up Invite — received from — %s %s`, name, lname)

	contents := []string{
		//fmt.Sprintf(`Dear %s %s`, name, lname),

		fmt.Sprintf(`%s %s has invited you to Sign Up (%s) for the GRAYLL App.`, name, lname, url),

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
