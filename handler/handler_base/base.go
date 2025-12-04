package handlerbase

import (
	"context"
	"sync"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/utility/logs"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type Operator[T, K any] interface {
	Name() string

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

		ForceReplyDirect bool
		SkipDone         bool
		Extra            map[string]any

		// TODO: 暂时没有用上，后续改造替换掉st、et的反复解析，搞成通用参数
		StartTime string
		EndTime   string
	}
)

func (m *BaseMetaData) GetExtra(key string) (any, bool) {
	if m.Extra == nil {
		m.Extra = make(map[string]any)
		return nil, false
	}
	val, ok := m.Extra[key]
	return val, ok
}

func (m *BaseMetaData) SetExtra(key string, val any) {
	if m.Extra == nil {
		m.Extra = make(map[string]any)
	}
	m.Extra[key] = val
}

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

func (op *OperatorBase[T, K]) Name() string {
	return "NotImplementBaseName"
}

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
	var span trace.Span
	p.Context, span = otel.LarkRobotOtelTracer.Start(p.Context, reflecting.GetCurrentFunc())
	defer span.End()

	for _, s := range p.stages {
		defer p.Defer()
		err = s.PreRun(p.Context, p.data, p.metaData)
		if err != nil {
			trace.SpanFromContext(p.Context).RecordError(err)
			if errors.Is(err, consts.ErrStageSkip) {
				logs.L().Ctx(p).Warn("Skipped pre run stage", zap.Error(err))
			} else {
				logs.L().Ctx(p).Error("Skipped pre run stage", zap.Error(err))
			}
			return
		}
		err = s.Run(p.Context, p.data, p.metaData)
		if err != nil {
			trace.SpanFromContext(p.Context).RecordError(err)
			if errors.Is(err, consts.ErrStageSkip) {
				logs.L().Ctx(p).Warn("run stage skipped", zap.Error(err))
			} else {
				logs.L().Ctx(p).Error("run stage skipped", zap.Error(err))
			}
			return
		}
		err = s.PostRun(p.Context, p.data, p.metaData)
		if err != nil {
			trace.SpanFromContext(p.Context).RecordError(err)
			if errors.Is(err, consts.ErrStageSkip) {
				logs.L().Ctx(p).Warn("post run stage skipped", zap.Error(err))
			} else {
				logs.L().Ctx(p).Error("post run stage skipped", zap.Error(err))
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
	wg := sync.WaitGroup{}
	wg.Go(func() { p.RunStages() })
	wg.Go(func() { p.RunParallelStages() })
	wg.Wait()
}

// RunParallelStages  运行并行处理阶段
//
//	@receiver p
//	@param ctx
//	@param event
//	@return error
func (p *Processor[T, K]) RunParallelStages() error {
	var span trace.Span
	p.Context, span = otel.LarkRobotOtelTracer.Start(p.Context, reflecting.GetCurrentFunc())
	defer span.End()

	wg := &sync.WaitGroup{}
	errorChan := make(chan error, len(p.parrallelStages))
	for _, operator := range p.parrallelStages {
		wg.Add(1)
		go func(op Operator[T, K]) {
			defer p.Defer()
			var err error
			defer func() {
				if err != nil && !errors.Is(err, consts.ErrStageSkip) {
					errorChan <- err
				}
				wg.Done()
			}()
			err = op.PreRun(p, p.data, p.metaData)
			if err != nil {
				if errors.Is(err, consts.ErrStageSkip) {
					logs.L().Ctx(p).Info("Skipped pre run stage", zap.Error(err))
				} else {
					trace.SpanFromContext(p.Context).RecordError(err)
					logs.L().Ctx(p).Error("pre run stage error", zap.Error(err))
				}
				return
			}

			logs.L().Ctx(p).Info("Run Handler", zap.String("handler", reflecting.GetFunctionName(op.Run)))
			err = op.Run(p.Context, p.data, p.metaData)
			if err != nil {
				if errors.Is(err, consts.ErrStageSkip) {
					logs.L().Ctx(p).Info("run stage skipped", zap.Error(err))
				} else {
					trace.SpanFromContext(p.Context).RecordError(err)
					logs.L().Ctx(p).Error("run stage error", zap.Error(err))
				}
				return
			}
			err = op.PostRun(p.Context, p.data, p.metaData)
			if err != nil {
				trace.SpanFromContext(p.Context).RecordError(err)
				if errors.Is(err, consts.ErrStageSkip) {
					logs.L().Ctx(p).Info("post run stage skipped", zap.Error(err))
				} else {
					logs.L().Ctx(p).Error("post run stage error", zap.Error(err))
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
			logs.L().Ctx(p).Warn("error in parallel stages", zap.Error(err))
		}
	}
	return mergedErr
}
