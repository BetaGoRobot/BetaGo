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
	defer func() {
		if e := recover(); e != nil {
			log.Println(a, e)
		}
	}()
	status, err := ajson.JSONPath(body, "$.action")
	if err != nil {
		log.Println(err.Error())
		return
	}
	workflowRun, err := ajson.JSONPath(body, "$.workflow_run")
	if err != nil {
		log.Println(err.Error())
		return
	}
	path, err := workflowRun[0].JSONPath("@.path")
	if err != nil {
		log.Println(err.Error())
		return
	}
	name, err := workflowRun[0].JSONPath("@.name")
	if err != nil {
		log.Println(err.Error())
		return
	}
	headCommit, err := workflowRun[0].JSONPath("@.head_commit")
	if err != nil {
		log.Println(err.Error())
		return
	}
	commitHash, err := headCommit[0].JSONPath("@.id")
	if err != nil {
		log.Println(err.Error())
		return
	}
	authorName, err := headCommit[0].JSONPath("@.author.name")
	if err != nil {
		log.Println(err.Error())
		return
	}
	htmlURL, err := workflowRun[0].JSONPath("@.html_url")
	if err != nil {
		log.Println(err.Error())
		return
	}
	conclusion, err := workflowRun[0].JSONPath("@.conclusion")
	if err != nil {
		log.Println(err.Error())
		return
	}
	msgID := utility.SendMessageWithTitle(
		GithubNotifyChan,
		"",
		"",
		strings.Join([]string{
			"**Commit:**\t" + fmt.Sprintf("[%s](%s)", commitHash[0].MustString()[:12], "https://github.com/BetaGoRobot/BetaGo/commit/"+commitHash[0].MustString()),
			"**Author:**\t" + fmt.Sprintf("[%s](https://github.com/%s)", authorName[0].MustString(), authorName[0].MustString()),
			"**Action:**\t" + fmt.Sprintf("[%s](%s)", name[0].String(), htmlURL[0].MustString()),
			"**Status:**\t`" + status[0].MustString() + "`-`" + conclusion[0].MustString() + "`",
		},
			"\n"),
		"New GitHub Action Event"+emoji.Pushpin.String(),
		ctx,
	)

	if conclusion[0].MustString() == "success" && status[0].MustString() == "completed" && path[0].MustString() == ".github/workflows/docker-image.yml" {
		// 是构建完成
		// 是正确的构建, 等待五秒后重启拉取最新的镜像
		go func() {
			time.Sleep(5 * time.Second)
			admin.RestartHandler(ctx, GithubNotifyChan, msgID, "")
		}()
	}
}

func Start() {
	r := router.New()
	r.GET("/", Index)
	r.POST("/webhook", WebHookHandler)
	log.Fatal(fasthttp.ListenAndServe(":8899", r.Handler))
}
