package packany

import (
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"unsafe"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

func toArgs(abiArgs abi.Arguments, input any) ([]any, error) {
	values := make([]any, 0, len(abiArgs))
	for i, arg := range abiArgs {
		v, ok := toArg(arg.Name, arg.Type, input)
		if !ok {
			return nil, fmt.Errorf("failed toArgs: %d, %s, %s", i, arg.Name, arg.Type.String())
		}
		values = append(values, v)
	}
	return values, nil
}

const (
	skipped     = true
	elemOfSlice = ""
)

func toArg(name string, abiTyp abi.Type, input any) (any, bool) {
	value, ok := toValue(name, input)
	if !ok {
		return nil, skipped
	}
	if value.Kind() == reflect.Invalid {
		return nil, false
	}

	var (
		rawValue         = value.Interface()
		strVal, isString = rawValue.(string)
		isHex            = isString && has0xPrefix(strVal)
		base             = ternary(isHex, 16, 10)
		typ              = abiTyp.T
		size             = abiTyp.Size
	)
	switch typ {
	case abi.IntTy:
		return toAbiInt(rawValue, base, size)

	case abi.UintTy:
		return toAbiUint(rawValue, base, size)

	case abi.BoolTy:
		if s, ok := toString(rawValue, true); ok {
			b, err := strconv.ParseBool(s)
			return b, err == nil
		}
		b, ok := rawValue.(bool)
		return b, ok

	case abi.StringTy:
		s, ok := toString(rawValue, false)
		return s, ok && unsafe.Sizeof(strVal) <= 32

	case abi.AddressTy:
		switch v := rawValue.(type) {
		case common.Address:
			return v, true
		case *common.Address:
			return *v, true
		case string:
			return common.HexToAddress(v), common.IsHexAddress(v)
		case []byte:
			return common.BytesToAddress(v), len(v) == 20
		default:
			return nil, false
		}

	case abi.BytesTy:
		if bytes, ok := rawValue.([]byte); ok {
			return bytes, true
		}
		if s, ok := toString(rawValue, true); ok {
			return common.Hex2Bytes(s), true
		}
		return nil, false

	case abi.FixedBytesTy:
		return toFixedBytes(value, size)

	case abi.FunctionTy:
		if value.Kind() == reflect.Interface {
			value = value.Elem()
		}
		if value.Kind() != reflect.Array || value.Len() != 24 {
			return nil, false
		}
		return rawValue, true

	case abi.TupleTy:
		tuple, ok := toTuple(abiTyp, rawValue)
		if !ok {
			return nil, false
		}
		return tuple.Interface(), true

	case abi.SliceTy, abi.ArrayTy:
		elemTyp := abiTyp.Elem
		value = reflect.ValueOf(rawValue)
		if !(value.Kind() == reflect.Slice || value.Kind() == reflect.Array) {
			return nil, false
		}

		slice := reflect.MakeSlice(toSliceTyp(*elemTyp), value.Len(), value.Len())
		if elemTyp.T == abi.TupleTy {
			for i := 0; i < value.Len(); i++ {
				elem, ok := toTuple(*elemTyp, value.Index(i).Interface())
				if !ok {
					return nil, false
				}
				slice.Index(i).Set(elem)
			}
		} else {
			for i := 0; i < value.Len(); i++ {
				arg, ok := toArg(elemOfSlice, *elemTyp, value.Index(i).Interface())
				if !ok {
					return nil, false
				}
				slice.Index(i).Set(reflect.ValueOf(arg))
			}
		}
		return slice.Interface(), true

	case abi.FixedPointTy, abi.HashTy: // currently not used in go-ethereum@v1.13.11
		return nil, false
	default:
		return nil, false
	}
}

func toValue(name string, input any) (reflect.Value, bool) {
	value := reflect.ValueOf(input)
	if !value.IsValid() {
		return reflect.Value{}, false
	}
	if name == elemOfSlice {
		return value, true
	}

	switch value.Kind() {
	case reflect.Struct:
		value = value.FieldByName(abi.ToCamelCase(name))
	case reflect.Map:
		for _, key := range value.MapKeys() {
			if key.String() == name {
				value = value.MapIndex(key)
				break
			}
		}
	case reflect.Func, reflect.Chan: // unsupported types
		return reflect.Value{}, false
	case reflect.Pointer:
		if value.IsNil() {
			return reflect.Value{}, false
		}
		return toValue(name, reflect.Indirect(value))
	}
	return value, true
}

func toString(input any, trim0x bool) (string, bool) {
	var s string
	switch v := input.(type) {
	case string:
		s = v
	case *string:
		s = *v
	default:
		return "", false
	}
	if trim0x && has0xPrefix(s) {
		s = s[2:]
	}
	return s, true
}

func toFixedBytes(v reflect.Value, size int) (any, bool) {
	var (
		bytes    []byte
		ok       bool
		rawValue = v.Interface()
	)
	switch v.Kind() {
	case reflect.Array:
		return rawValue, v.Len() == size
	case reflect.Interface:
		elem := v.Elem()
		return rawValue, elem.Kind() == reflect.Array && elem.Len() == size
	case reflect.String:
		var s string
		s, ok = toString(rawValue, true)
		if !ok {
			return nil, false
		}
		bytes = common.Hex2Bytes(s)
	default:
		bytes, ok = rawValue.([]byte)
	}
	if !ok || len(bytes) != size {
		return nil, false
	}
	typ := reflect.ArrayOf(size, reflect.TypeOf(uint8(0)))
	arr := reflect.New(typ).Elem()
	reflect.Copy(arr, reflect.ValueOf(bytes))
	return arr.Interface(), true
}

func toInt64(x any, base, bitSize int) (int64, bool) {
	switch y := x.(type) {
	case string:
		if base == 16 && has0xPrefix(y) {
			y = y[2:]
		}
		signed, err := strconv.ParseInt(y, base, bitSize)
		return signed, err == nil
	case *big.Int:
		if y == nil {
			return 0, true
		}
		return y.Int64(), true
	case big.Int:
		return y.Int64(), true
	case uint:
		return int64(y), true
	case uint8:
		return int64(y), true
	case uint16:
		return int64(y), true
	case uint32:
		return int64(y), true
	case uint64:
		return int64(y), true
	case int:
		return int64(y), true
	case int8:
		return int64(y), true
	case int16:
		return int64(y), true
	case int32:
		return int64(y), true
	case int64:
		return y, true
	case float32:
		return int64(y), true
	case float64:
		return int64(y), true
	default:
		rv := reflect.ValueOf(y)
		if rv.Kind() == reflect.Pointer {
			return toInt64(reflect.Indirect(rv).Interface(), base, bitSize)
		}
		return 0, false
	}
}

func toUint64(x any, base, bitSize int) (uint64, bool) {
	switch y := x.(type) {
	case string:
		if base == 16 && has0xPrefix(y) {
			y = y[2:]
		}
		unsigned, err := strconv.ParseUint(y, base, bitSize)
		return unsigned, err == nil
	case *big.Int:
		if y == nil {
			return 0, true
		}
		return y.Uint64(), y.Sign() >= 0
	case uint:
		return uint64(y), true
	case uint8:
		return uint64(y), true
	case uint16:
		return uint64(y), true
	case uint32:
		return uint64(y), true
	case uint64:
		return y, true
	case int:
		return uint64(y), y >= 0
	case int8:
		return uint64(y), y >= 0
	case int16:
		return uint64(y), y >= 0
	case int32:
		return uint64(y), y >= 0
	case int64:
		return uint64(y), y >= 0
	case float32:
		return uint64(y), y >= 0
	case float64:
		return uint64(y), y >= 0
	default:
		rv := reflect.ValueOf(y)
		if rv.Kind() == reflect.Pointer {
			return toUint64(reflect.Indirect(rv).Interface(), base, bitSize)
		}
		return 0, false
	}
}

func toBigInt(x any, base int, signed bool) (*big.Int, bool) {
	switch y := x.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64,
		*int, *int8, *int16, *int32, *int64, *uint, *uint8, *uint16, *uint32, *uint64, *float32, *float64:
		if signed {
			if z, ok := toInt64(y, 10, 64); ok {
				return new(big.Int).SetInt64(z), true
			}
		}
		if z, ok := toUint64(y, 10, 64); ok {
			return new(big.Int).SetUint64(z), true
		}
	case string:
		if base == 16 && has0xPrefix(y) {
			y = y[2:]
		}
		return new(big.Int).SetString(y, base)
	case *big.Int:
		return y, true
	case big.Int:
		return &y, true
	}
	return nil, false
}

func toAbiInt(x any, base, bitSize int) (any, bool) {
	if bitSize > 64 {
		return toBigInt(x, base, true)
	}
	i64, ok := toInt64(x, base, bitSize)
	if !ok {
		return nil, false
	}

	max := int64(1<<(bitSize-1) - 1)
	min := int64(-1 << (bitSize - 1))
	if i64 > max || i64 < min {
		return nil, false
	}

	switch bitSize {
	case 8:
		return int8(i64), true
	case 16:
		return int16(i64), true
	case 32:
		return int32(i64), true
	case 64:
		return i64, true
	default:
		return new(big.Int).SetInt64(i64), true
	}
}

func toAbiUint(x any, base, bitSize int) (any, bool) {
	if bitSize > 64 {
		return toBigInt(x, base, false)
	}
	ui64, ok := toUint64(x, base, bitSize)
	if !ok {
		return nil, false
	}

	max := uint64(1<<bitSize - 1)
	if ui64 > max {
		return nil, false
	}

	switch bitSize {
	case 8:
		return uint8(ui64), true
	case 16:
		return uint16(ui64), true
	case 32:
		return uint32(ui64), true
	case 64:
		return ui64, true
	default:
		return new(big.Int).SetUint64(ui64), true
	}
}

func toTuple(abiTyp abi.Type, input any) (reflect.Value, bool) {
	values := make([]any, 0, len(abiTyp.TupleElems))
	for i, subTyp := range abiTyp.TupleElems {
		val, ok := toArg(abiTyp.TupleRawNames[i], *subTyp, input)
		if !ok {
			return reflect.Value{}, false
		}
		values = append(values, val)
	}

	tuple := reflect.New(abiTyp.TupleType).Elem()
	for i, v := range values {
		tuple.Field(i).Set(reflect.ValueOf(v))
	}
	return tuple, true
}

func toSliceTyp(abiTyp abi.Type) reflect.Type {
	if abiTyp.T == abi.TupleTy {
		return reflect.SliceOf(abiTyp.TupleType)
	} else {
		return reflect.SliceOf(abiTyp.GetType())
	}
}
