package handlerbase

import (
	"context"
	"sync"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/kevinmatthe/zaplog"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/trace"
)

type Operator[T, K any] interface {
	PreRun(context.Context, *T, *K) error
	Run(context.Context, *T, *K) error
	PostRun(context.Context, *T, *K) error

	MetaInit() *K
}

type (
	OperatorBase[T, K any] struct{}
	BaseMetaData           struct {
		Refresh     bool
		IsCommand   bool
		MainCommand string
		TraceID     string
		UserID      string
	}
)

type (
	ProcPanicFunc[T, K any] func(context.Context, error, *T, *K)
	ProcDeferFunc[T, K any] func(context.Context, *T, *K)
	MetaInitFunc[T, K any]  func(context.Context) (*K, error)
	Processor[T, K any]     struct {
		context.Context

		needBreak       bool
		data            *T
		metaData        *K
		stages          []Operator[T, K]
		parrallelStages []Operator[T, K]
		onPanicFn       ProcPanicFunc[T, K]
		deferFn         []ProcDeferFunc[T, K]
	}
)

func (op *OperatorBase[T, K]) PreRun(context.Context, *T, *K) error {
	return nil
}

func (op *OperatorBase[T, K]) Run(context.Context, *T, *K) error {
	return nil
}

func (op *OperatorBase[T, K]) PostRun(context.Context, *T, *K) error {
	return nil
}

func (op *OperatorBase[T, K]) MetaInit() *K {
	return new(K)
}

func (p *Processor[T, K]) WithCtx(ctx context.Context) *Processor[T, K] {
	p.Context = ctx
	return p
}

func (p *Processor[T, K]) OnPanic(fn ProcPanicFunc[T, K]) *Processor[T, K] {
	p.onPanicFn = fn
	return p
}

func (p *Processor[T, K]) WithDefer(fns ...ProcDeferFunc[T, K]) *Processor[T, K] {
	p.deferFn = append(p.deferFn, fns...)
	return p
}

func (p *Processor[T, K]) WithEvent(event *T) *Processor[T, K] {
	p.data = event
	return p
}

func (p *Processor[T, K]) Clean() *Processor[T, K] {
	p.data = nil
	p.Context = nil
	return p
}

func (p *Processor[T, K]) Defer() {
	if err := recover(); err != nil {
		if p.onPanicFn != nil {
			p.onPanicFn(p.Context, err.(error), p.data, p.metaData)
		}
	}
}

// AddStages  添加处理阶段
//
//	@receiver p
//	@param stage
//	@return *Processor[T]
func (p *Processor[T, K]) AddStages(stage Operator[T, K]) *Processor[T, K] {
	p.stages = append(p.stages, stage)
	return p
}

// AddParallelStages  添加并行处理阶段
//
//	@receiver p
//	@param stage
//	@return *Processor[T]
func (p *Processor[T, K]) AddParallelStages(stage Operator[T, K]) *Processor[T, K] {
	p.parrallelStages = append(p.parrallelStages, stage)
	return p
}

// RunStages  运行处理阶段
//
//	@receiver p
//	@param ctx
//	@param event
func (p *Processor[T, K]) RunStages() (err error) {
	for _, s := range p.stages {
		defer p.Defer()
		err = s.PreRun(p.Context, p.data, p.metaData)
		if err != nil {
			trace.SpanFromContext(p.Context).RecordError(err)
			if errors.Is(err, consts.ErrStageSkip) {
				log.Zlog.Warn("pre run stage skipped: ", zaplog.Error(err))
			} else {
				log.Zlog.Error("pre run stage skipped: ", zaplog.Error(err))
			}
			return
		}
		err = s.Run(p.Context, p.data, p.metaData)
		if err != nil {
			trace.SpanFromContext(p.Context).RecordError(err)
			if errors.Is(err, consts.ErrStageSkip) {
				log.Zlog.Warn("run stage skipped: ", zaplog.Error(err))
			} else {
				log.Zlog.Error("run stage skipped: ", zaplog.Error(err))
			}
			return
		}
		err = s.PostRun(p.Context, p.data, p.metaData)
		if err != nil {
			trace.SpanFromContext(p.Context).RecordError(err)
			if errors.Is(err, consts.ErrStageSkip) {
				log.Zlog.Warn("post run stage skipped: ", zaplog.Error(err))
			} else {
				log.Zlog.Error("post run stage skipped: ", zaplog.Error(err))
			}
			return
		}
	}
	return
}

// Run  运行
//
//	@receiver p
//	@param ctx
//	@param event
func (p *Processor[T, K]) Run() {
	p.metaData = new(K)
	for _, fn := range p.deferFn {
		if fn != nil {
			defer fn(p.Context, p.data, p.metaData)
		}
	}

	go p.RunStages()
	go p.RunParallelStages()
}

// RunParallelStages  运行并行处理阶段
//
//	@receiver p
//	@param ctx
//	@param event
//	@return error
func (p *Processor[T, K]) RunParallelStages() error {
	wg := &sync.WaitGroup{}
	errorChan := make(chan error, len(p.parrallelStages))

	for _, operator := range p.parrallelStages {
		wg.Add(1)
		go func(op Operator[T, K]) {
			defer p.Defer()
			var err error
			defer func() {
				errorChan <- err
				wg.Done()
			}()
			err = op.PreRun(p.Context, p.data, p.metaData)
			if err != nil {
				trace.SpanFromContext(p.Context).RecordError(err)
				if errors.Is(err, consts.ErrStageSkip) {
					log.Zlog.Warn("pre run stage skipped: ", zaplog.Error(err))
				} else {
					log.Zlog.Error("pre run stage skipped: ", zaplog.Error(err))
				}
				return
			}

			err = op.Run(p.Context, p.data, p.metaData)
			if err != nil {
				trace.SpanFromContext(p.Context).RecordError(err)
				if errors.Is(err, consts.ErrStageSkip) {
					log.Zlog.Warn("run stage skipped: ", zaplog.Error(err))
				} else {
					log.Zlog.Error("run stage skipped: ", zaplog.Error(err))
				}
				return
			}
			err = op.PostRun(p.Context, p.data, p.metaData)
			if err != nil {
				trace.SpanFromContext(p.Context).RecordError(err)
				if errors.Is(err, consts.ErrStageSkip) {
					log.Zlog.Warn("post run stage skipped: ", zaplog.Error(err))
				} else {
					log.Zlog.Error("post run stage skipped: ", zaplog.Error(err))
				}
				return
			}
		}(operator)
	}
	go func() {
		defer close(errorChan)
		wg.Wait()
	}()
	var mergedErr error
	for err := range errorChan {
		if err != nil {
			mergedErr = errors.Wrap(mergedErr, err.Error())
			log.Zlog.Warn("error in parallel stages", zaplog.Error(err))
		}
	}
	return mergedErr
}
