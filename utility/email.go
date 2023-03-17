package utility

import (
	"fmt"
	"net/smtp"
	"os"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/jordan-wright/email"
	"github.com/kevinmatthe/zaplog"
)

const (
	netEaseEmailURL         = "smtp.163.com"
	netEaseEmailURLWithPort = "smtp.163.com:25"
	netEaseEmailAddress     = "betagonotfi@163.com"
)

var (
	netEaseEmailSecret = os.Getenv("MAIL_SECRET")
	recevierEmailList  = make([]string, 0)
)

// SendEmail  is the function to send email
func SendEmail(Subject string, Body string) {
	getReceieverEmailList()
	em := email.NewEmail()
	em.From = fmt.Sprintf("%s <%s>", betagovar.RobotName, netEaseEmailAddress)
	em.To = recevierEmailList
	em.Subject = Subject
	em.Text = []byte(Body)
	err := em.Send(netEaseEmailURLWithPort, smtp.PlainAuth("", netEaseEmailAddress, netEaseEmailSecret, netEaseEmailURL))
	if err != nil {
		ZapLogger.DPanic("Send email failed", zaplog.Error(err))
	}
	ZapLogger.Info("Send email success")
}

// SendQRCodeMail is the function to send QRCode mail
//
//	@param qrimg
//	@return error
func SendQRCodeMail(qrimg string) error {
	getReceieverEmailList()
	em := email.NewEmail()
	em.From = fmt.Sprintf("%s <%s>", betagovar.RobotName, netEaseEmailAddress)
	em.To = recevierEmailList
	em.Subject = "网易云登陆提醒"
	em.AttachFile(qrimg)
	err := em.Send(netEaseEmailURLWithPort, smtp.PlainAuth("", netEaseEmailAddress, netEaseEmailSecret, netEaseEmailURL))
	if err != nil {
		return err
	}
	os.Remove(qrimg)
	return nil
}

// getReceieverEmailList is the function to get recevier email list
func getReceieverEmailList() {
	globalDBConn.Table("betago.alert_lists").Find(&recevierEmailList)
}
