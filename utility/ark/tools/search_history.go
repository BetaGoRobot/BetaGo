package tools

import (
	"context"
	"strings"

	"github.com/BetaGoRobot/BetaGo/utility/ark/embedding"
	"github.com/BetaGoRobot/BetaGo/utility/history"
	"github.com/bytedance/sonic"
)

type SearchArgs struct {
	Keywords  string `json:"keywords"`
	TopK      int    `json:"top_k"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	UserID    string `json:"user_id"`
}

func HybridSearch(ctx context.Context, meta *FunctionCallMeta, argStr string) (res any, err error) {
	args := &SearchArgs{}
	err = sonic.UnmarshalString(argStr, &args)
	if err != nil {
		return
	}
	res, err = history.HybridSearch(ctx,
		history.HybridSearchRequest{
			QueryText: strings.Split(args.Keywords, ","),
			TopK:      args.TopK,
			UserID:    args.UserID,
			ChatID:    meta.ChatID,
			StartTime: args.StartTime,
			EndTime:   args.EndTime,
		}, embedding.EmbeddingText)
	if err != nil {
		return
	}
	return
}
