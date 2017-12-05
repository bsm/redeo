package resp

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Append implements ResponseWriter
func (w *bufioW) Append(v interface{}) error {
	switch v := v.(type) {
	case nil:
		w.AppendNil()
	case CustomResponse:
		v.AppendTo(w)
	case error:
		msg := v.Error()
		if !strings.HasPrefix(msg, "ERR ") {
			msg = "ERR " + msg
		}
		w.AppendError(msg)
	case bool:
		if v {
			w.AppendInt(1)
		} else {
			w.AppendInt(0)
		}
	case int:
		w.AppendInt(int64(v))
	case int8:
		w.AppendInt(int64(v))
	case int16:
		w.AppendInt(int64(v))
	case int32:
		w.AppendInt(int64(v))
	case int64:
		w.AppendInt(v)
	case uint:
		w.AppendInt(int64(v))
	case uint8:
		w.AppendInt(int64(v))
	case uint16:
		w.AppendInt(int64(v))
	case uint32:
		w.AppendInt(int64(v))
	case uint64:
		w.AppendInt(int64(v))
	case string:
		w.AppendBulkString(v)
	case []byte:
		w.AppendBulk(v)
	case CommandArgument:
		w.AppendBulk(v)
	case float32:
		w.AppendInlineString(strconv.FormatFloat(float64(v), 'f', -1, 32))
	case float64:
		w.AppendInlineString(strconv.FormatFloat(v, 'f', -1, 64))
	default:
		switch reflect.TypeOf(v).Kind() {
		case reflect.Slice:
			s := reflect.ValueOf(v)

			w.AppendArrayLen(s.Len())
			for i := 0; i < s.Len(); i++ {
				w.Append(s.Index(i).Interface())
			}
		case reflect.Map:
			s := reflect.ValueOf(v)

			w.AppendArrayLen(s.Len() * 2)
			for _, key := range s.MapKeys() {
				w.Append(key.Interface())
				w.Append(s.MapIndex(key).Interface())
			}
		default:
			return fmt.Errorf("resp: unsupported type %T", v)
		}
	}
	return nil
}
