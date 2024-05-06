package gotify

import (
	"context"
	"net/http"
	"net/url"
	"os"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/gotify/go-api-client/v2/auth"
	"github.com/gotify/go-api-client/v2/client"
	"github.com/gotify/go-api-client/v2/client/message"
	"github.com/gotify/go-api-client/v2/gotify"
	"github.com/gotify/go-api-client/v2/models"
	"github.com/kevinmatthe/zaplog"
)

var (
	gotifyURL           = os.Getenv("GOTIFY_URL")
	applicationToken    = os.Getenv("GOTIFY_APPLICATION_TOKEN")
	tokenParsed         = auth.TokenAuth(applicationToken)
	DefaultGotifyClient *client.GotifyREST
)

func init() {
	gotifyURLParsed, err := url.Parse(gotifyURL)
	if err != nil {
		panic("error parsing url for gotify" + err.Error())
	}
	DefaultGotifyClient = gotify.NewClient(gotifyURLParsed, &http.Client{})
}

func SendMessage(ctx context.Context, title, msg string, priority int) {
	if title == "" {
		title = "BetaGo Notification"
	}
	title = "[" + betagovar.BotIdentifier + "]" + title
	params := message.NewCreateMessageParams()
	params.Body = &models.MessageExternal{
		Title:    title,
		Message:  msg,
		Priority: priority,
		Extras: map[string]interface{}{
			"client::display": map[string]string{"contentType": "text/markdown"},
		},
	}

	_, err := DefaultGotifyClient.Message.CreateMessage(params, tokenParsed)
	if err != nil {
		log.ZapLogger.Error("Could not send message %v", zaplog.Error(err))
		return
	}
	log.ZapLogger.Info("Gotify Message Sent!")
}
