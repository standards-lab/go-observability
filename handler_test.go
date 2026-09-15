package observability_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"go.opentelemetry.io/otel/trace"

	"github.com/standards-lab/go-observability"
)

// recordingHandler is a slog.Handler double that keeps the last record it
// received, the attributes bound through WithAttrs, and the groups opened
// through WithGroup. WithAttrs and WithGroup return copies that share the
// same capture, so a test inspects one value whichever derived handler the
// record arrived through.
type recordingHandler struct {
	level   slog.Level
	capture *capture
}

type capture struct {
	record  slog.Record
	handled bool
	attrs   []slog.Attr
	groups  []string
}

func newRecordingHandler(level slog.Level) *recordingHandler {
	return &recordingHandler{level: level, capture: &capture{}}
}

func (h *recordingHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *recordingHandler) Handle(_ context.Context, record slog.Record) error {
	h.capture.record = record
	h.capture.handled = true
	return nil
}

func (h *recordingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.capture.attrs = append(h.capture.attrs, attrs...)
	return &recordingHandler{level: h.level, capture: h.capture}
}

func (h *recordingHandler) WithGroup(name string) slog.Handler {
	h.capture.groups = append(h.capture.groups, name)
	return &recordingHandler{level: h.level, capture: h.capture}
}

var (
	testTraceID = trace.TraceID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
	testSpanID  = trace.SpanID{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
)

// spanContext returns a context carrying a valid, sampled span context built
// from the test identifiers.
func spanContext() context.Context {
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    testTraceID,
		SpanID:     testSpanID,
		TraceFlags: trace.FlagsSampled,
	})
	return trace.ContextWithSpanContext(context.Background(), sc)
}

// recordAttrs collects a record's attributes into a map keyed by attribute
// name, so a test asserts on presence and value without walking the record.
func recordAttrs(record slog.Record) map[string]slog.Value {
	attrs := make(map[string]slog.Value, record.NumAttrs())
	record.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value
		return true
	})
	return attrs
}

func newRecord(msg string) slog.Record {
	record := slog.NewRecord(time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC), slog.LevelInfo, msg, 0)
	record.AddAttrs(slog.String("existing", "value"))
	return record
}

func assertTraceAttrs(t *testing.T, record slog.Record) {
	t.Helper()
	attrs := recordAttrs(record)
	if got, want := attrs["trace_id"].String(), testTraceID.String(); got != want {
		t.Errorf("trace_id = %q, want %q", got, want)
	}
	if got, want := attrs["span_id"].String(), testSpanID.String(); got != want {
		t.Errorf("span_id = %q, want %q", got, want)
	}
}

func TestTraceHandlerAddsTraceAttributesUnderValidSpan(t *testing.T) {
	t.Parallel()
	next := newRecordingHandler(slog.LevelDebug)
	handler := observability.NewTraceHandler(next)

	if err := handler.Handle(spanContext(), newRecord("hello")); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if !next.capture.handled {
		t.Fatal("wrapped handler did not receive the record")
	}

	got := next.capture.record
	assertTraceAttrs(t, got)
	if got.NumAttrs() != 3 {
		t.Errorf("NumAttrs = %d, want 3 (existing, trace_id, span_id)", got.NumAttrs())
	}
	if v := recordAttrs(got)["existing"].String(); v != "value" {
		t.Errorf("existing attr = %q, want %q", v, "value")
	}
}

func TestTraceHandlerPassesRecordThroughWithoutSpan(t *testing.T) {
	t.Parallel()
	next := newRecordingHandler(slog.LevelDebug)
	handler := observability.NewTraceHandler(next)
	in := newRecord("hello")

	if err := handler.Handle(context.Background(), in); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if !next.capture.handled {
		t.Fatal("wrapped handler did not receive the record")
	}

	got := next.capture.record
	attrs := recordAttrs(got)
	if _, ok := attrs["trace_id"]; ok {
		t.Error("trace_id present on a record logged without a span")
	}
	if _, ok := attrs["span_id"]; ok {
		t.Error("span_id present on a record logged without a span")
	}
	if got.Message != in.Message {
		t.Errorf("Message = %q, want %q", got.Message, in.Message)
	}
	if got.Level != in.Level {
		t.Errorf("Level = %v, want %v", got.Level, in.Level)
	}
	if !got.Time.Equal(in.Time) {
		t.Errorf("Time = %v, want %v", got.Time, in.Time)
	}
	if got.NumAttrs() != 1 {
		t.Errorf("NumAttrs = %d, want 1", got.NumAttrs())
	}
	if v := attrs["existing"].String(); v != "value" {
		t.Errorf("existing attr = %q, want %q", v, "value")
	}
}

func TestTraceHandlerWithAttrsKeepsCorrelating(t *testing.T) {
	t.Parallel()
	next := newRecordingHandler(slog.LevelDebug)
	logger := slog.New(observability.NewTraceHandler(next)).With("key", "val")

	logger.InfoContext(spanContext(), "hello")

	if !next.capture.handled {
		t.Fatal("wrapped handler did not receive the record")
	}
	assertTraceAttrs(t, next.capture.record)
	if len(next.capture.attrs) != 1 || next.capture.attrs[0].Key != "key" || next.capture.attrs[0].Value.String() != "val" {
		t.Errorf("bound attrs = %v, want [key=val]", next.capture.attrs)
	}
}

func TestTraceHandlerWithGroupKeepsCorrelating(t *testing.T) {
	t.Parallel()
	next := newRecordingHandler(slog.LevelDebug)
	logger := slog.New(observability.NewTraceHandler(next)).WithGroup("req")

	logger.InfoContext(spanContext(), "hello")

	if !next.capture.handled {
		t.Fatal("wrapped handler did not receive the record")
	}
	assertTraceAttrs(t, next.capture.record)
	if len(next.capture.groups) != 1 || next.capture.groups[0] != "req" {
		t.Errorf("opened groups = %v, want [req]", next.capture.groups)
	}
}

func TestTraceHandlerEnabledDelegates(t *testing.T) {
	t.Parallel()
	handler := observability.NewTraceHandler(newRecordingHandler(slog.LevelWarn))
	ctx := context.Background()

	for _, level := range []slog.Level{slog.LevelDebug, slog.LevelInfo} {
		if handler.Enabled(ctx, level) {
			t.Errorf("Enabled(%v) = true, want false", level)
		}
	}
	for _, level := range []slog.Level{slog.LevelWarn, slog.LevelError} {
		if !handler.Enabled(ctx, level) {
			t.Errorf("Enabled(%v) = false, want true", level)
		}
	}
}
