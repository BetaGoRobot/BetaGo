// Package larkhandler  存放lark消息的处理handler
//
//	@update 2024-07-16 09:52:36
package larkhandler

import (
	"context"
	"sync"

	"github.com/BetaGoRobot/BetaGo/handler"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/kevinmatthe/zaplog"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/pkg/errors"
)

var _ handler.BotMsgProcessor = &LarkMsgProcessor{}

type (
	// LarkMsgProcessor struct  消息处理器
	//	@update 2024-07-16 09:52:28
	LarkMsgProcessor struct {
		context.Context
		event           *larkim.P2MessageReceiveV1
		stages          []LarkMsgOperator
		parrallelStages []LarkMsgOperator
	}

	LarkMsgOperatorBase struct{}
	// LarkMsgOperator interface  算子接口
	//	@update 2024-07-16 09:52:20
	LarkMsgOperator interface {
		PreRun(context.Context, *larkim.P2MessageReceiveV1) error
		Run(context.Context, *larkim.P2MessageReceiveV1) error
		PostRun(context.Context, *larkim.P2MessageReceiveV1) error
	}
)

func (op *LarkMsgOperatorBase) PreRun(context.Context, *larkim.P2MessageReceiveV1) error {
	return nil
}

func (op *LarkMsgOperatorBase) Run(context.Context, *larkim.P2MessageReceiveV1) error {
	return nil
}

func (op *LarkMsgOperatorBase) PostRun(context.Context, *larkim.P2MessageReceiveV1) error {
	return nil
}

func (p *LarkMsgProcessor) WithCtx(ctx context.Context) *LarkMsgProcessor {
	p.Context = ctx
	return p
}

func (p *LarkMsgProcessor) WithEvent(event *larkim.P2MessageReceiveV1) *LarkMsgProcessor {
	p.event = event
	return p
}

func (p *LarkMsgProcessor) Clean() *LarkMsgProcessor {
	p.event = nil
	p.Context = nil
	return p
}

// AddStages  添加处理阶段
//
//	@receiver p
//	@param stage
//	@return *LarkMsgProcessor
func (p *LarkMsgProcessor) AddStages(stage LarkMsgOperator) *LarkMsgProcessor {
	p.stages = append(p.stages, stage)
	return p
}

// AddParallelStages  添加并行处理阶段
//
//	@receiver p
//	@param stage
//	@return *LarkMsgProcessor
func (p *LarkMsgProcessor) AddParallelStages(stage LarkMsgOperator) *LarkMsgProcessor {
	p.parrallelStages = append(p.parrallelStages, stage)
	return p
}

// RunStages  运行处理阶段
//
//	@receiver p
//	@param ctx
//	@param event
func (p *LarkMsgProcessor) RunStages() (err error) {
	defer larkutils.RecoverMsg(p.Context, *p.event.Event.Message.MessageId)

	for _, s := range p.stages {
		err = s.PreRun(p.Context, p.event)
		if err != nil {
			if errors.Is(err, ErrStageSkip) {
				log.ZapLogger.Warn("pre run stage skipped: ", zaplog.Error(err))
			} else {
				log.ZapLogger.Error("pre run stage skipped: ", zaplog.Error(err))
			}
			return
		}
		err = s.Run(p.Context, p.event)
		if err != nil {
			if errors.Is(err, ErrStageSkip) {
				log.ZapLogger.Warn("run stage skipped: ", zaplog.Error(err))
			} else {
				log.ZapLogger.Error("run stage skipped: ", zaplog.Error(err))
			}
			return
		}
		err = s.PostRun(p.Context, p.event)
		if err != nil {
			if errors.Is(err, ErrStageSkip) {
				log.ZapLogger.Warn("post run stage skipped: ", zaplog.Error(err))
			} else {
				log.ZapLogger.Error("post run stage skipped: ", zaplog.Error(err))
			}
			return
		}
	}
	return
}

// RunParallelStages  运行并行处理阶段
//
//	@receiver p
//	@param ctx
//	@param event
//	@return error
func (p *LarkMsgProcessor) RunParallelStages() error {
	defer larkutils.RecoverMsg(p.Context, *p.event.Event.Message.MessageId)

	wg := sync.WaitGroup{}
	errorChan := make(chan error, len(p.parrallelStages))

	for _, operator := range p.parrallelStages {
		wg.Add(1)
		go func(op LarkMsgOperator) {
			var err error
			defer func() {
				errorChan <- err
				wg.Done()
			}()
			err = op.PreRun(p.Context, p.event)
			if err != nil {
				if errors.Is(err, ErrStageSkip) {
					log.ZapLogger.Warn("pre run stage skipped: ", zaplog.Error(err))
				} else {
					log.ZapLogger.Error("pre run stage skipped: ", zaplog.Error(err))
				}
				return
			}
			err = op.Run(p.Context, p.event)
			if err != nil {
				if errors.Is(err, ErrStageSkip) {
					log.ZapLogger.Warn("run stage skipped: ", zaplog.Error(err))
				} else {
					log.ZapLogger.Error("run stage skipped: ", zaplog.Error(err))
				}
				return
			}
			err = op.PostRun(p.Context, p.event)
			if err != nil {
				if errors.Is(err, ErrStageSkip) {
					log.ZapLogger.Warn("post run stage skipped: ", zaplog.Error(err))
				} else {
					log.ZapLogger.Error("post run stage skipped: ", zaplog.Error(err))
				}
				return
			}
		}(operator)
	}
	wg.Wait()
	close(errorChan)
	var mergedErr error
	for err := range errorChan {
		if err != nil {
			mergedErr = errors.Wrap(mergedErr, err.Error())
			log.ZapLogger.Warn("error in parallel stages", zaplog.Error(err))
		}
	}
	return mergedErr
}
