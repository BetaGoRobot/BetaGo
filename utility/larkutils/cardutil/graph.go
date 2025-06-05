package cardutil

import (
	"context"

	"github.com/BetaGoRobot/BetaGo/cts"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/templates"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/BetaGo/utility/vadvisor"
	"github.com/BetaGoRobot/go_utils/reflecting"
)

type CardBuilderGraph[X cts.ValidType, Y cts.Numeric] struct {
	*CardBuilderBase
	graph *vadvisor.MultiSeriesLineGraph[X, Y]
}

// func  NewCardBuildGraphHelper
//
//	@update 2025-06-05 13:30:47
func NewCardBuildGraphHelper[X cts.ValidType, Y cts.Numeric](graph *vadvisor.MultiSeriesLineGraph[X, Y]) *CardBuilderGraph[X, Y] {
	return &CardBuilderGraph[X, Y]{
		CardBuilderBase: NewCardBuildHelper(),
		graph:           graph,
	}
}

func (h *CardBuilderGraph[X, Y]) SetTitle(title string) *CardBuilderGraph[X, Y] {
	h.CardBuilderBase.SetTitle(title)
	return h
}

func (h *CardBuilderGraph[X, Y]) SetSubTitle(subTitle string) *CardBuilderGraph[X, Y] {
	h.SetSubTitle(subTitle)
	return h
}

func (h *CardBuilderGraph[X, Y]) SetContent(text string) *CardBuilderGraph[X, Y] {
	h.SetContent(text)
	return h
}

func (h *CardBuilderGraph[X, Y]) Build(ctx context.Context) *templates.TemplateCardContent {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	cardContent := templates.NewCardContent(ctx, templates.NormalCardGraphReplyTemplate)
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
