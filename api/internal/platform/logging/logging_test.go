package logging

import (
	"context"
	"encoding/json"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestModuleLoggerEmitsBoundedStructuredFields(t *testing.T) {
	var output []byte
	core := zapcore.NewCore(zapcore.NewJSONEncoder(zapcore.EncoderConfig{MessageKey: "msg"}), zapcore.AddSync(&byteBuffer{value: &output}), zapcore.DebugLevel)
	logger := NewModuleLogger(zap.New(core), "orders")
	logger.Info(context.Background(), "order.loaded", Field{Key: "order_id", Value: "order-1"}, Field{Key: "password", Value: "must-not-log"}, Field{Key: "Bad Key", Value: "must-not-log"})

	var record map[string]any
	if err := json.Unmarshal(output, &record); err != nil {
		t.Fatal(err)
	}
	if record["module"] != "orders" || record["order_id"] != "order-1" {
		t.Fatalf("unexpected structured fields: %#v", record)
	}
	if _, ok := record["password"]; ok {
		t.Fatal("sensitive field was emitted")
	}
	if _, ok := record["Bad Key"]; ok {
		t.Fatal("invalid field was emitted")
	}
}

func TestValidFieldRejectsUnsupportedKeys(t *testing.T) {
	for _, field := range []Field{
		{Key: "", Value: "value"},
		{Key: "1order", Value: "value"},
		{Key: "client_secret", Value: "value"},
		{Key: "access_token", Value: "value"},
		{Key: "service", Value: "value"},
		{Key: "requestId", Value: "value"},
	} {
		if validField(field) {
			t.Fatalf("validField(%q) = true", field.Key)
		}
	}
}

type byteBuffer struct{ value *[]byte }

func (b *byteBuffer) Write(value []byte) (int, error) {
	*b.value = append(*b.value, value...)
	return len(value), nil
}

func (b *byteBuffer) Sync() error { return nil }
