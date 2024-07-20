package base

import (
	"context"
	"sync"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/kevinmatthe/zaplog"
	"github.com/pkg/errors"
)

type Operator[T any] interface {
	PreRun(context.Context, *T) error
	Run(context.Context, *T) error
	PostRun(context.Context, *T) error
}

type OperatorBase[T any] struct{}

type (
	ProcessorDeferFunc[T any] func(context.Context, any, *T)
	Processor[T any]          struct {
		context.Context
		needBreak       bool
		data            *T
		stages          []Operator[T]
		parrallelStages []Operator[T]
		deferFunc       ProcessorDeferFunc[T]
	}
)

func (op *OperatorBase[T]) PreRun(context.Context, *T) error {
	return nil
}

func (op *OperatorBase[T]) Run(context.Context, *T) error {
	return nil
}

func (op *OperatorBase[T]) PostRun(context.Context, *T) error {
	return nil
}

func (p *Processor[T]) WithCtx(ctx context.Context) *Processor[T] {
	p.Context = ctx
	return p
}

func (p *Processor[T]) WithDefer(fn ProcessorDeferFunc[T]) *Processor[T] {
	p.deferFunc = fn
	return p
}

func (p *Processor[T]) WithEvent(event *T) *Processor[T] {
	p.data = event
	return p
}

func (p *Processor[T]) Clean() *Processor[T] {
	p.data = nil
	p.Context = nil
	return p
}

func (p *Processor[T]) Defer() {
	if err := recover(); err != nil {
		p.deferFunc(p.Context, err, p.data)
	}
}

// AddStages  添加处理阶段
//
//	@receiver p
//	@param stage
//	@return *Processor[T]
func (p *Processor[T]) AddStages(stage Operator[T]) *Processor[T] {
	p.stages = append(p.stages, stage)
	return p
}

// AddParallelStages  添加并行处理阶段
//
//	@receiver p
//	@param stage
//	@return *Processor[T]
func (p *Processor[T]) AddParallelStages(stage Operator[T]) *Processor[T] {
	p.parrallelStages = append(p.parrallelStages, stage)
	return p
}

// RunStages  运行处理阶段
//
//	@receiver p
//	@param ctx
//	@param event
func (p *Processor[T]) RunStages() (err error) {
	for _, s := range p.stages {
		err = s.PreRun(p.Context, p.data)
		if err != nil {
			if errors.Is(err, consts.ErrStageSkip) {
				log.ZapLogger.Warn("pre run stage skipped: ", zaplog.Error(err))
			} else {
				log.ZapLogger.Error("pre run stage skipped: ", zaplog.Error(err))
			}
			return
		}
		err = s.Run(p.Context, p.data)
		if err != nil {
			if errors.Is(err, consts.ErrStageSkip) {
				log.ZapLogger.Warn("run stage skipped: ", zaplog.Error(err))
			} else {
				log.ZapLogger.Error("run stage skipped: ", zaplog.Error(err))
			}
			return
		}
		err = s.PostRun(p.Context, p.data)
		if err != nil {
			if errors.Is(err, consts.ErrStageSkip) {
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
func (p *Processor[T]) RunParallelStages() error {
	wg := sync.WaitGroup{}
	errorChan := make(chan error, len(p.parrallelStages))

	for _, operator := range p.parrallelStages {
		wg.Add(1)
		go func(op Operator[T]) {
			defer p.Defer()
			var err error
			defer func() {
				errorChan <- err
				wg.Done()
			}()
			err = op.PreRun(p.Context, p.data)
			if err != nil {
				if errors.Is(err, consts.ErrStageSkip) {
					log.ZapLogger.Warn("pre run stage skipped: ", zaplog.Error(err))
				} else {
					log.ZapLogger.Error("pre run stage skipped: ", zaplog.Error(err))
				}
				return
			}

			err = op.Run(p.Context, p.data)
			if err != nil {
				if errors.Is(err, consts.ErrStageSkip) {
					log.ZapLogger.Warn("run stage skipped: ", zaplog.Error(err))
				} else {
					log.ZapLogger.Error("run stage skipped: ", zaplog.Error(err))
				}
				return
			}
			err = op.PostRun(p.Context, p.data)
			if err != nil {
				if errors.Is(err, consts.ErrStageSkip) {
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
