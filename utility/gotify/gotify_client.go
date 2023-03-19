package gotify

import (
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/gotify/go-api-client/v2/auth"
	"github.com/gotify/go-api-client/v2/client"
	"github.com/gotify/go-api-client/v2/client/message"
	"github.com/gotify/go-api-client/v2/gotify"
	"github.com/gotify/go-api-client/v2/models"
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

func SendMessage(title, msg string, priority int) {
	if title == "" {
		title = "BetaGo Notification"
	}
	params := message.NewCreateMessageParams()
	params.Body = &models.MessageExternal{
		Title:    title,
		Message:  msg,
		Priority: priority,
		Extras:   map[string]interface{}{"client::display": map[string]string{"contentType": "text/markdown"}},
	}

	_, err := DefaultGotifyClient.Message.CreateMessage(params, tokenParsed)
	if err != nil {
		log.Panicf("Could not send message %v", err)
		return
	}
	log.Println("Message Sent!")
}
