package domain

import (
	"fmt"

	"sparkbridge/pkg/spbproto"
)

// InferDataType maps native Go values to Sparkplug B data types.
func InferDataType(val any) (spbproto.DataType, error) {
	switch val.(type) {
	case int32:
		return spbproto.DataType_DATA_TYPE_INT32, nil
	case int64:
		return spbproto.DataType_DATA_TYPE_INT64, nil
	case float32:
		return spbproto.DataType_DATA_TYPE_FLOAT, nil
	case float64:
		return spbproto.DataType_DATA_TYPE_DOUBLE, nil
	case string:
		return spbproto.DataType_DATA_TYPE_STRING, nil
	case bool:
		return spbproto.DataType_DATA_TYPE_BOOLEAN, nil
	case []byte:
		return spbproto.DataType_DATA_TYPE_BYTES, nil
	default:
		return spbproto.DataType_DATA_TYPE_UNKNOWN, fmt.Errorf("unsupported data type %T", val)
	}
}
