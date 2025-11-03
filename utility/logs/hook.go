// internal/logging/hook.go
package logs

import (
	"context"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// SpanEventHook 是一个 zerolog.Hook，它将日志事件转换为 OpenTelemetry 的 Span Event
// 并将其附加到当前活动的 Span 上。
type SpanEventHook struct{}

// Run 方法在每次记录日志时被 zerolog 调用。
func (h SpanEventHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	ctx := e.GetCtx()
	if ctx == nil {
		ctx = context.Background()
	}

	span := trace.SpanFromContext(ctx)
	// 仅当存在正在录制的活动 Span 时才继续
	if !span.IsRecording() {
		return
	}

	// 创建一个属性切片来保存日志数据
	attrs := make([]attribute.KeyValue, 0)
	attrs = append(attrs, attribute.String("log.severity", level.String()))
	attrs = append(attrs, attribute.String("log.message", msg))

	// **重要限制**: zerolog 的高性能设计使其无法在 Hook 中以结构化方式
	// 轻松地提取已添加到事件中的字段（如.Str("key", "val")）。
	// 此处展示的方法仅转换了日志级别和消息。
	// 要包含所有结构化字段，需要更复杂的、可能基于反射的解决方案，
	// 或者使用专门为 OpenTelemetry 设计的日志库/桥接器。

	// 将日志作为 Span Event 添加
	span.AddEvent("log", trace.WithAttributes(attrs...))

	// 检查 span 是否有效。如果 context 中没有 span，
	// SpanFromContext 会返回一个无操作的 span，其 SpanContext 是无效的。
	// IsValid() 是安全检查的关键，避免在日志中添加空的 ID。
	if span.SpanContext().IsValid() {
		e.Str("trace_id", span.SpanContext().TraceID().String())
		e.Str("span_id", span.SpanContext().SpanID().String())
	}
}
