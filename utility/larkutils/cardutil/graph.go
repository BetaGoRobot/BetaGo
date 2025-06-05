package cardutil

import (
	"context"

	"github.com/BetaGoRobot/BetaGo/utility/larkutils/templates"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
)

type CardBuilderGraph struct {
	*CardBuilderBase
	graph any
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

func (h *CardBuilderGraph) Build(ctx context.Context) *templates.TemplateCardContent {
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
