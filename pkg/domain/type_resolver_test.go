package domain

import (
	"testing"

	"sparkbridge/pkg/spbproto"
)

func TestInferDataType(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		want    spbproto.DataType
		wantErr bool
	}{
		{name: "int32", value: int32(1), want: spbproto.DataType_DATA_TYPE_INT32},
		{name: "int64", value: int64(1), want: spbproto.DataType_DATA_TYPE_INT64},
		{name: "float32", value: float32(1), want: spbproto.DataType_DATA_TYPE_FLOAT},
		{name: "float64", value: float64(1), want: spbproto.DataType_DATA_TYPE_DOUBLE},
		{name: "string", value: "x", want: spbproto.DataType_DATA_TYPE_STRING},
		{name: "bool", value: true, want: spbproto.DataType_DATA_TYPE_BOOLEAN},
		{name: "bytes", value: []byte("x"), want: spbproto.DataType_DATA_TYPE_BYTES},
		{name: "unsupported", value: struct{}{}, want: spbproto.DataType_DATA_TYPE_UNKNOWN, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InferDataType(tt.value)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
	}
}
