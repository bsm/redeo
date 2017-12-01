package resp

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const (
	errMsgNilPtr = "destination pointer is nil"
	errMsgNotPtr = "destination not a pointer"
)

// Scan attempts to scan responses into given values
func (b *bufioR) Scan(vv ...interface{}) error {
	for _, v := range vv {
		if err := b.scan(v); err != nil {
			return err
		}
	}
	return nil
}

func (b *bufioR) scan(dst interface{}) error {
	pt, err := b.PeekType()
	if err != nil {
		return err
	}

	if pt == TypeError {
		src, err := b.ReadError()
		if err != nil {
			return err
		}
		return fmt.Errorf(`resp: server error %q`, src)
	}

	if scn, ok := dst.(Scannable); ok {
		return scn.ScanResponse(pt, b)
	}

	switch pt {
	case TypeArray:
		sz, err := b.ReadArrayLen()
		if err != nil {
			return err
		}
		return b.scanArray(dst, sz)
	case TypeNil:
		if err := b.ReadNil(); err != nil {
			return err
		}
		return scanNil(dst)
	case TypeInt:
		src, err := b.ReadInt()
		if err != nil {
			return err
		}
		return scanInt(dst, src)
	case TypeInline:
		src, err := b.ReadInlineString()
		if err != nil {
			return err
		}
		return scanString(dst, src)
	case TypeBulk:
		if v, ok := dst.(*[]byte); ok {
			src, err := b.ReadBulk(nil)
			if err != nil {
				return err
			}
			return assignBytes(v, src)
		}

		src, err := b.ReadBulkString()
		if err != nil {
			return err
		}
		return scanString(dst, src)
	default:
		return errBadResponseType
	}
}

func (b *bufioR) scanArray(dst interface{}, sz int) error {
	dpv := reflect.ValueOf(dst)

	// skip array if nil is passed
	if dpv.Kind() == reflect.Invalid && dst == nil {
		for i := 0; i < sz; i++ {
			if err := b.scan(nil); err != nil {
				return err
			}
		}
		return nil
	}
	if dpv.Kind() != reflect.Ptr {
		return scanErrf(dst, errMsgNotPtr)
	}
	if dpv.IsNil() {
		return scanErrf(dst, errMsgNilPtr)
	}

	dv := reflect.Indirect(dpv)
	switch dv.Kind() {
	case reflect.Slice:
		if dv.Len() < sz {
			nv := reflect.MakeSlice(dv.Type(), sz, sz)
			reflect.Copy(nv, dv)
			dv.Set(nv)
		}

		var err error
		for i := 0; i < sz; i++ {
			val := dv.Index(i)
			if val.Kind() != reflect.Ptr && val.CanAddr() {
				val = val.Addr()
			}
			if e := b.scan(val.Interface()); e != nil && err == nil {
				err = e
			}
		}
		return err
	case reflect.Map:
		if sz%2 != 0 {
			break
		}

		dt := dv.Type()
		if dv.IsNil() {
			dv.Set(reflect.MakeMap(dt))
		}

		kt, vt := dt.Key(), dt.Elem()

		var err error
		for i := 0; i < sz; i += 2 {
			key, val := reflect.New(kt), reflect.New(vt)
			if e := b.Scan(key.Interface(), val.Interface()); e != nil && err == nil {
				err = e
			} else {
				dv.SetMapIndex(key.Elem(), val.Elem())
			}
		}
		return err
	}

	return scanErrf(dst, "unsupported conversion from array[%d]", sz)
}

func scanErrf(dst interface{}, format string, vv ...interface{}) error {
	vv = append([]interface{}{dst}, vv...)
	return fmt.Errorf("resp: error on Scan into %T: "+format, vv...)
}

func scanNil(dst interface{}) error {
	switch w := dst.(type) {
	case *interface{}:
		if w == nil {
			return scanErrf(dst, errMsgNilPtr)
		}
		*w = nil
		return nil
	case *[]byte:
		if w == nil {
			return scanErrf(dst, errMsgNilPtr)
		}
		*w = nil
		return nil
	case nil:
		return nil
	}

	return scanValue(dst, nil)
}

func scanString(dst interface{}, src string) error {
	switch v := dst.(type) {
	case *string:
		if v == nil {
			return scanErrf(dst, errMsgNilPtr)
		}
		*v = src
		return nil
	case *[]byte:
		if v == nil {
			return scanErrf(dst, errMsgNilPtr)
		}
		*v = []byte(src)
		return nil
	case *bool:
		if v == nil {
			return scanErrf(dst, errMsgNilPtr)
		}
		if src == "1" || src == "0" || strings.ToUpper(src) == "OK" {
			*v = src != "0"
			return nil
		}
	case *int:
		if v == nil {
			return scanErrf(dst, errMsgNilPtr)
		}
		if n, err := strconv.ParseInt(src, 10, 64); err == nil {
			*v = int(n)
			return nil
		}
	case *int8:
		if v == nil {
			return scanErrf(dst, errMsgNilPtr)
		}
		if n, err := strconv.ParseInt(src, 10, 8); err == nil {
			*v = int8(n)
			return nil
		}
	case *int16:
		if v == nil {
			return scanErrf(dst, errMsgNilPtr)
		}
		if n, err := strconv.ParseInt(src, 10, 16); err == nil {
			*v = int16(n)
			return nil
		}
	case *int32:
		if v == nil {
			return scanErrf(dst, errMsgNilPtr)
		}
		if n, err := strconv.ParseInt(src, 10, 32); err == nil {
			*v = int32(n)
			return nil
		}
	case *int64:
		if v == nil {
			return scanErrf(dst, errMsgNilPtr)
		}
		if n, err := strconv.ParseInt(src, 10, 64); err == nil {
			*v = int64(n)
			return nil
		}
	case *uint:
		if v == nil {
			return scanErrf(dst, errMsgNilPtr)
		}
		if n, err := strconv.ParseUint(src, 10, 64); err == nil {
			*v = uint(n)
			return nil
		}
	case *uint8:
		if v == nil {
			return scanErrf(dst, errMsgNilPtr)
		}
		if n, err := strconv.ParseUint(src, 10, 8); err == nil {
			*v = uint8(n)
			return nil
		}
	case *uint16:
		if v == nil {
			return scanErrf(dst, errMsgNilPtr)
		}
		if n, err := strconv.ParseUint(src, 10, 16); err == nil {
			*v = uint16(n)
			return nil
		}
	case *uint32:
		if v == nil {
			return scanErrf(dst, errMsgNilPtr)
		}
		if n, err := strconv.ParseUint(src, 10, 32); err == nil {
			*v = uint32(n)
			return nil
		}
	case *uint64:
		if v == nil {
			return scanErrf(dst, errMsgNilPtr)
		}
		if n, err := strconv.ParseUint(src, 10, 64); err == nil {
			*v = uint64(n)
			return nil
		}
	case *float32:
		if v == nil {
			return scanErrf(dst, errMsgNilPtr)
		}
		if n, err := strconv.ParseFloat(src, 32); err == nil {
			*v = float32(n)
			return nil
		}
	case *float64:
		if v == nil {
			return scanErrf(dst, errMsgNilPtr)
		}
		if n, err := strconv.ParseFloat(src, 64); err == nil {
			*v = float64(n)
			return nil
		}
	case nil:
		return nil
	}
	return scanValue(dst, src)
}

func scanInt(dst interface{}, src int64) error {
	switch v := dst.(type) {
	case *string:
		if v == nil {
			return scanErrf(dst, errMsgNilPtr)
		}
		*v = strconv.FormatInt(src, 10)
		return nil
	case *[]byte:
		if v == nil {
			return scanErrf(dst, errMsgNilPtr)
		}
		*v = []byte(strconv.FormatInt(src, 10))
		return nil
	case *bool:
		if v == nil {
			return scanErrf(dst, errMsgNilPtr)
		}
		if src == 0 || src == 1 {
			*v = src == 1
			return nil
		}
	case *int:
		if v == nil {
			return scanErrf(dst, errMsgNilPtr)
		}
		*v = int(src)
		return nil
	case *int8:
		if v == nil {
			return scanErrf(dst, errMsgNilPtr)
		}
		*v = int8(src)
		return nil
	case *int16:
		if v == nil {
			return scanErrf(dst, errMsgNilPtr)
		}
		*v = int16(src)
		return nil
	case *int32:
		if v == nil {
			return scanErrf(dst, errMsgNilPtr)
		}
		*v = int32(src)
		return nil
	case *int64:
		if v == nil {
			return scanErrf(dst, errMsgNilPtr)
		}
		*v = int64(src)
		return nil
	case *uint:
		if v == nil {
			return scanErrf(dst, errMsgNilPtr)
		}
		*v = uint(src)
		return nil
	case *uint8:
		if v == nil {
			return scanErrf(dst, errMsgNilPtr)
		}
		*v = uint8(src)
		return nil
	case *uint16:
		if v == nil {
			return scanErrf(dst, errMsgNilPtr)
		}
		*v = uint16(src)
		return nil
	case *uint32:
		if v == nil {
			return scanErrf(dst, errMsgNilPtr)
		}
		*v = uint32(src)
		return nil
	case *uint64:
		if v == nil {
			return scanErrf(dst, errMsgNilPtr)
		}
		*v = uint64(src)
		return nil
	case *float32:
		if v == nil {
			return scanErrf(dst, errMsgNilPtr)
		}
		*v = float32(src)
		return nil
	case *float64:
		if v == nil {
			return scanErrf(dst, errMsgNilPtr)
		}
		*v = float64(src)
		return nil
	case nil:
		return nil
	}
	return scanValue(dst, src)
}

func scanValue(dst, src interface{}) error {
	dpv := reflect.ValueOf(dst)
	if dpv.Kind() != reflect.Ptr {
		return scanErrf(dst, errMsgNotPtr)
	}
	if dpv.IsNil() {
		return scanErrf(dst, errMsgNilPtr)
	}

	dv := reflect.Indirect(dpv)
	sv := reflect.ValueOf(src)

	// check if directly assignable
	if sv.IsValid() && sv.Type().AssignableTo(dv.Type()) {
		dv.Set(sv)
		return nil
	}

	// check if same kind and convertable
	if dv.Kind() == sv.Kind() && sv.Type().ConvertibleTo(dv.Type()) {
		dv.Set(sv.Convert(dv.Type()))
		return nil
	}

	return scanErrf(dst, "unsupported conversion from %#v", src)
}

func assignBytes(v *[]byte, src []byte) error {
	if v == nil {
		return scanErrf(v, errMsgNotPtr)
	}

	*v = src
	return nil
}
