package utility

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"runtime/debug"
	"strings"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/BetaGoRobot/BetaGo/utility/gotify"
	"github.com/BetaGoRobot/BetaGo/utility/jaeger_client"
	"github.com/enescakir/emoji"
	jsoniter "github.com/json-iterator/go"
	"go.opentelemetry.io/otel/attribute"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// CollectPanic is the function to collect panic
func CollectPanic(ctx context.Context, kookCtx interface{}, TargetID, QuoteID, UserID string) {
	if err := recover(); err != nil {
		ctx, span := jaeger_client.BetaGoCommandTracer.Start(ctx, GetCurrentFunc())
		span.SetAttributes(attribute.Key("Panic Stack").String(removeSensitiveInfo(debug.Stack())))
		defer span.End()

		JSONStr := ForceMarshalJSON(ctx)
		SendEmail("Panic-Collected!", fmt.Sprintf("%v\n%s", string(debug.Stack()), JSONStr))
		// // 测试频道不用脱敏
		SendMessageWithTitle(betagovar.TestChanID, "", "",
			fmt.Sprintf(emoji.ExclamationMark.String()+"发生Panic, 请保存此ID提供给开发者\nTraceID: "+span.SpanContext().TraceID().String()),
			fmt.Sprintf("%s Panic-Collected!",
				emoji.Warning.String()))
		gotify.SendMessage(
			fmt.Sprintf("%s Panic-Collected!",
				emoji.Warning.String()),
			strings.ReplaceAll(fmt.Sprintf("SourceChannelID: `%s`\nErrorMsg: `%s`\n```go\n%s```\nRecord:\n```json\n%s\n```",
				TargetID, err, removeSensitiveInfo(debug.Stack()), JSONStr), "\n", "\n\n"),
			7)
		if TargetID != betagovar.TestChanID {
			SendMessageWithTitle(TargetID, "", "",
				fmt.Sprintf(emoji.ExclamationMark.String()+"发生Panic, 请保存此ID提供给开发者\nTraceID: "+span.SpanContext().TraceID().String()),
				fmt.Sprintf("%s Panic-Collected!",
					emoji.Warning.String()))
		}
		SugerLogger.Errorf("=====Panic====== %s", string(debug.Stack()))
	}
}

func removeSensitiveInfo(stack []byte) string {
	r := bytes.NewReader(stack)
	buf := make([]byte, 0)
	var res []byte
	for {
		b, err := r.ReadByte()
		if err == io.EOF {
			break
		}
		if b == '\n' {
			if leftIndex, rightIndex := strings.LastIndex(string(buf), "("), strings.LastIndex(string(buf), ")"); leftIndex+1 != rightIndex {
				if leftIndex != -1 && rightIndex != -1 {
					temp := buf[:leftIndex+1]
					temp = append(temp, buf[rightIndex:]...)
					buf = temp
				}
			}
			newD := strings.Split(string(buf), " ")
			newD = strings.Split(string(strings.Join(newD[:len(newD)-1], "")), "/")
			if len(newD) > 1 {
				res = append(res, []byte("\t/"+strings.Join(newD[len(newD)-2:], "/"))...)
			} else {
				res = append(res, buf...)
			}
			res = append(res, '\n')
			buf = make([]byte, 0)
		} else {
			buf = append(buf, b)
		}
	}
	return string(res)
}

// ForceMarshalJSON is the function to force marshal json
//
//	@param v
//	@return string
func ForceMarshalJSON(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "    ")

	return string(b)
}
