package cardutil

import (
	"context"
	"time"

	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/templates"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
)

type CardBuilderGraph struct {
	*CardBuilderBase
	graph  any
	st, et time.Time
}

// func  NewCardBuildGraphHelper
//
//	@update 2025-06-05 13:30:47
func NewCardBuildGraphHelper(graph any) *CardBuilderGraph {
	return &CardBuilderGraph{
		CardBuilderBase: NewCardBuildHelper(),
		graph:           graph,
	}
}

func (h *CardBuilderGraph) SetTitle(title string) *CardBuilderGraph {
	h.CardBuilderBase.SetTitle(title)
	return h
}

func (h *CardBuilderGraph) SetSubTitle(subTitle string) *CardBuilderGraph {
	h.SetSubTitle(subTitle)
	return h
}

func (h *CardBuilderGraph) SetContent(text string) *CardBuilderGraph {
	h.SetContent(text)
	return h
}

func (h *CardBuilderGraph) SetStartTime(t time.Time) *CardBuilderGraph {
	h.st = t
	return h
}

func (h *CardBuilderGraph) SetEndTime(t time.Time) *CardBuilderGraph {
	h.et = t
	return h
}

func (h *CardBuilderGraph) Build(ctx context.Context) *templates.TemplateCardContent {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	cardContent := templates.NewCardContent(ctx, templates.NormalCardGraphReplyTemplate)
	if !h.st.IsZero() && !h.et.IsZero() {
		cardContent.
			AddVariable("start_time", h.st.In(utility.UTCPlus8Loc()).Format("2006-01-02 15:04")).
			AddVariable("end_time", h.et.In(utility.UTCPlus8Loc()).Format("2006-01-02 15:04"))
	}
	return cardContent.
		AddVariable(
			"title", h.Title,
		).
		AddVariable(
			"subtitle", h.SubTitle,
		).
		AddVariable(
			"content", h.Content,
		).
		AddVariable(
			"graph", h.graph,
		)
}
