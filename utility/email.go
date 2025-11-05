package utility

import (
	"fmt"
	"net/smtp"
	"os"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/utility/logs"
	"github.com/jordan-wright/email"
	"gorm.io/gorm"
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
	GetReceieverEmailList(consts.GlobalDBConn)
	em := email.NewEmail()
	em.From = fmt.Sprintf("%s <%s>", consts.RobotName, netEaseEmailAddress)
	em.To = recevierEmailList
	em.Subject = Subject
	em.Text = []byte(Body)
	err := em.Send(netEaseEmailURLWithPort, smtp.PlainAuth("", netEaseEmailAddress, netEaseEmailSecret, netEaseEmailURL))
	if err != nil {
		logs.L.Error().Err(err).Msg("Send email failed")
	}
	logs.L.Info().Msg("Send email success")
}

// SendQRCodeMail is the function to send QRCode mail
//
//	@param qrimg
//	@return error
func SendQRCodeMail(qrimg string) error {
	GetReceieverEmailList(consts.GlobalDBConn)
	em := email.NewEmail()
	em.From = fmt.Sprintf("%s <%s>", consts.RobotName, netEaseEmailAddress)
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

// GetReceieverEmailList is the function to get recevier email list
func GetReceieverEmailList(dbConn *gorm.DB) {
	dbConn.Table("betago.alert_lists").Find(&recevierEmailList)
}
