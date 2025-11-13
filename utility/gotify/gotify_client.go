package gotify

import (
	"context"
	"net/http"
	"net/url"
	"os"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/utility/logs"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/gotify/go-api-client/v2/auth"
	"github.com/gotify/go-api-client/v2/client"
	"github.com/gotify/go-api-client/v2/client/message"
	"github.com/gotify/go-api-client/v2/gotify"
	"github.com/gotify/go-api-client/v2/models"
	"go.uber.org/zap"
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
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, "SendMessage")
	defer span.End()
	logs.L().Ctx(ctx).Info("SendMessage...", zap.String("traceID", span.SpanContext().TraceID().String()))

	if title == "" {
		title = "BetaGo Notification"
	}
	title = "[" + consts.BotIdentifier + "]" + title
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
		logs.L().Ctx(ctx).Error("Could not send message", zap.Error(err))
		return
	}
	logs.L().Ctx(ctx).Info("Gotify Message Sent!")
}
