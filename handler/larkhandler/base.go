// Package larkhandler  存放lark消息的处理handler
//
//	@update 2024-07-16 09:52:36
package larkhandler

import (
	"context"
	"sync"

	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/kevinmatthe/zaplog"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/pkg/errors"
)

type (
	// LarkMsgProcessor struct  消息处理器
	//	@update 2024-07-16 09:52:28
	LarkMsgProcessor struct {
		context.Context
		stages          []LarkMsgOperator
		parrallelStages []LarkMsgOperator
	}

	// LarkMsgOperator interface  算子接口
	//	@update 2024-07-16 09:52:20
	LarkMsgOperator interface {
		PreRun(context.Context, *larkim.P2MessageReceiveV1) error
		Run(context.Context, *larkim.P2MessageReceiveV1) error
		PostRun(context.Context, *larkim.P2MessageReceiveV1) error
	}
)

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
func (p *LarkMsgProcessor) RunStages(ctx context.Context, event *larkim.P2MessageReceiveV1) {
	for _, s := range p.stages {
		s.PreRun(ctx, event)
		s.Run(ctx, event)
		s.PostRun(ctx, event)
	}
}

// RunParallelStages  运行并行处理阶段
//
//	@receiver p
//	@param ctx
//	@param event
//	@return error
func (p *LarkMsgProcessor) RunParallelStages(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
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
			err = op.PreRun(ctx, event)
			if err != nil {
				err = errors.Wrap(err, "error in pre run")
				return
			}
			err = op.Run(ctx, event)
			if err != nil {
				err = errors.Wrap(err, "error in run")
				return
			}
			err = op.PostRun(ctx, event)
			if err != nil {
				err = errors.Wrap(err, "error in post run")
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
			log.ZapLogger.Error("error in parallel stages", zaplog.Error(err))
		}
	}
	return mergedErr
}
