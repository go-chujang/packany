package packany

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

func PackAny(abi abi.ABI, name string, input any) (packed []byte, err error) {
	method, exist := abi.Methods[name]
	if !exist {
		return nil, errors.New("method not found")
	}
	args, err := toArgs(method.Inputs, input)
	if err != nil {
		return nil, err
	}
	return abi.Pack(method.Name, args...)
}

func PackArgs(abi abi.ABI, name string, args ...any) (packed []byte, err error) {
	method, exist := abi.Methods[name]
	if !exist {
		return nil, errors.New("method not found")
	}
	cnt := len(method.Inputs)
	if cnt != len(args) {
		return nil, fmt.Errorf("expected %d arguments, but got %d", cnt, len(args))
	}
	parsed := make([]any, 0, cnt)
	for i, arg := range method.Inputs {
		v, err := toArg(arg, args[i])
		if err != nil {
			return nil, err
		}
		parsed = append(parsed, v)
	}
	return abi.Pack(method.Name, parsed...)
}

func ToFunctionTy(id []byte, contractAddress common.Address) (funtionTy [24]byte, ok bool) {
	if len(id) != 4 {
		return [24]byte{}, false
	}
	functionTyBytes := append(contractAddress.Bytes(), id...)
	copy(funtionTy[:], functionTyBytes[0:24])
	return funtionTy, true
}
