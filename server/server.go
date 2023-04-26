package server

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/BetaGoRobot/BetaGo/commandHandler/admin"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/enescakir/emoji"
	"github.com/fasthttp/router"
	"github.com/spyzhov/ajson"
	"github.com/valyala/fasthttp"
)

func Index(ctx *fasthttp.RequestCtx) {
	ctx.WriteString("Welcome!")
}

func WebHookHandler(ctx *fasthttp.RequestCtx) {
	const GithubNotifyChan = "6841156803258223"
	var a map[string]interface{}

	if string(ctx.Request.Header.Peek("X-Github-Event")) != "workflow_run" {
		fmt.Fprintf(ctx, "OK")
		return
	}
	body := ctx.Request.Body()
	json.Unmarshal(body, &a)
	status, _ := ajson.JSONPath(body, "$.action")
	workflowRun, _ := ajson.JSONPath(body, "$.workflow_run")
	path, _ := workflowRun[0].JSONPath("@.path")
	name, _ := workflowRun[0].JSONPath("@.name")
	headCommit, _ := workflowRun[0].JSONPath("@.head_commit")
	commitHash, _ := headCommit[0].JSONPath("@.id")
	authorName, _ := headCommit[0].JSONPath("@.author.name")
	htmlURL, _ := workflowRun[0].JSONPath("@.html_url")
	conclusion, _ := workflowRun[0].JSONPath("@.conclusion")
	utility.SendMessageWithTitle(
		GithubNotifyChan,
		"",
		"",
		strings.Join([]string{
			"**Commit:**\t`" + commitHash[0].MustString()[:12] + "`",
			"**Author:**\t" + fmt.Sprintf("[%s](https://github.com/%s)", authorName[0].MustString(), authorName[0].MustString()),
			"**Action:**\t`" + name[0].MustString() + "`",
			"**Stage:**\t`" + status[0].MustString() + "`",
			"**Conclusion:**\t`" + conclusion[0].MustString() + "`",
			"[ActionURL](" + htmlURL[0].MustString() + ")",
		},
			"\n"),
		"New GitHub Action Event"+emoji.Pushpin.String(),
		ctx,
	)

	if conclusion[0].MustString() == "success" && status[0].MustString() == "completed" && path[0].MustString() == ".github/workflows/docker-image.yml" {
		// 是构建完成
		// 是正确的构建, 等待五秒后重启拉取最新的镜像
		time.Sleep(5 * time.Second)
		admin.RestartHandler(ctx, "", "", "")
	}
}

func Start() {
	r := router.New()
	r.GET("/", Index)
	r.POST("/webhook", WebHookHandler)
	go log.Fatal(fasthttp.ListenAndServe(":8899", r.Handler))
}
