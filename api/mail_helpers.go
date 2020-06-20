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
