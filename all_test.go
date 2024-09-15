package packany

import (
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

const (
	methodArgTuple    = "argTuple"
	methodArgCallback = "argCallback"
)

func Test_All(t *testing.T) {
	output_anypack1, err := PackAny(abijson, methodArgTuple, input_anypack)
	if err != nil {
		t.Fatal(err)
	}
	output_anypack2, err := PackById(abijson, abijson.Methods[methodArgCallback].ID, map[string]interface{}{
		"x":        [32]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31},
		"y":        big.NewInt(10),
		"callback": fnty1,
	})
	if err != nil {
		t.Fatal(err)
	}

	output_goether1, err := abijson.Pack(methodArgTuple, input_goether)
	if err != nil {
		t.Fatal(err)
	}
	output_goether2, err := abijson.Pack(methodArgCallback,
		[32]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31},
		big.NewInt(10),
		fnty1)
	if err != nil {
		t.Fatal(err)
	}

	if string(output_anypack1) != string(output_goether1) {
		t.Fail()
	}
	if string(output_anypack2) != string(output_goether2) {
		t.Fail()
	}
}

func Benchmark_anypack(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, err := PackAny(abijson, methodArgTuple, input_anypack); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_goether(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, err := abijson.Pack(methodArgTuple, input_goether); err != nil {
			b.Fatal(err)
		}
	}
}

type (
	tuple1 struct {
		X256     *big.Int         `json:"x256"`
		Y24      *big.Int         `json:"y24"`
		Z256     *big.Int         `json:"z256"`
		Str      string           `json:"str"`
		Account  common.Address   `json:"account"`
		Accounts []common.Address `json:"accounts"`
		Tuples   []tuple2         `json:"tuples"`
	}
	tuple2 struct {
		B   bool     `json:"b"`
		Bs  []byte   `json:"bs"`
		Fbs [24]byte `json:"fbs"`
	}
)

var (
	contractAddress = "0xD33258f4B6d2a1136B2A3e771B51a2F7D593Be42"
	abijson, _      = abi.JSON(strings.NewReader(abiJsonString))

	input_anypack = map[string]interface{}{
		"t": map[string]interface{}{
			"x256":    "10",
			"y24":     "0x5",
			"z256":    big.NewInt(-10),
			"str":     "foobar",
			"account": common.HexToAddress("0x0000000000000000000000000000000000000001"),
			"accounts": []interface{}{
				"0x0000000000000000000000000000000000000002",
				common.HexToAddress("0x0000000000000000000000000000000000000003"),
				common.HexToAddress("0x0000000000000000000000000000000000000004").Bytes()},
			"tuples": []interface{}{
				map[string]interface{}{
					"b":   true,
					"bs":  []byte("foo"),
					"fbs": [24]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23}},
				map[string]string{
					"b":   "0x0",
					"bs":  "0x11",
					"fbs": "0x000102030405060708091011121314151617181920212223"},
			},
		},
	}
	input_goether = tuple1{
		X256:    big.NewInt(10),
		Y24:     big.NewInt(5),
		Z256:    big.NewInt(-10),
		Str:     "foobar",
		Account: common.HexToAddress("0x0000000000000000000000000000000000000001"),
		Accounts: []common.Address{
			common.HexToAddress("0x0000000000000000000000000000000000000002"),
			common.HexToAddress("0x0000000000000000000000000000000000000003"),
			common.HexToAddress("0x0000000000000000000000000000000000000004"),
		},
		Tuples: []tuple2{
			{
				B:   true,
				Bs:  []byte{102, 111, 111},
				Fbs: [24]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23}},
			{
				B:   false,
				Bs:  []byte{17},
				Fbs: [24]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 32, 33, 34, 35}},
		},
	}
	fnty1, _ = ToFunctionTy(abijson.Methods["fnty_1"].ID, common.HexToAddress(contractAddress))
)

const (
	abiJsonString = `[
	{
		"inputs": [
			{
				"components": [
					{"internalType": "uint256", "name": "x256", "type": "uint256"},
					{"internalType": "uint24", "name": "y24", "type": "uint24"},
					{"internalType": "int256", "name": "z256","type": "int256"},
					{"internalType": "string","name": "str","type": "string"},
					{"internalType": "address","name": "account","type": "address"},
					{"internalType": "address[]","name": "accounts","type": "address[]"},
					{
						"components": [
							{"internalType": "bool","name": "b","type": "bool"},
							{"internalType": "bytes","name": "bs","type": "bytes"},
							{"internalType": "bytes24","name": "fbs","type": "bytes24"}
						],
						"internalType": "struct TupleTest.tuple2[]","name": "tuples","type": "tuple[]"
					}
				],
				"internalType": "struct TupleTest.tuple1","name": "t","type": "tuple"
			}
		],
		"name": "argTuple",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"internalType": "bytes32","name": "x","type": "bytes32"},
			{"internalType": "uint256","name": "y","type": "uint256"},
			{"internalType": "function (uint256) external returns (uint256)","name": "callback","type": "function"}
		],
		"name": "argCallback",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"internalType": "uint256","name": "x","type": "uint256"}],
		"name": "fnty_1",
		"outputs": [{"internalType": "uint256","name": "","type": "uint256"}],
		"stateMutability": "pure",
		"type": "function"
	},
	{
		"inputs": [{"internalType": "uint256","name": "x","type": "uint256"}],
		"name": "fnty_2",
		"outputs": [{"internalType": "uint256","name": "","type": "uint256"}],
		"stateMutability": "pure",
		"type": "function"
	},
	{
		"inputs": [{"internalType": "uint256", "name": "idx", "type": "uint256"}],
		"name": "check",
		"outputs": [
			{
				"components": [
					{"internalType": "uint256", "name": "x256", "type": "uint256"},
					{"internalType": "uint24", "name": "y24", "type": "uint24"},
					{"internalType": "int256", "name": "z256","type": "int256"},
					{"internalType": "string","name": "str","type": "string"},
					{"internalType": "address","name": "account","type": "address"},
					{"internalType": "address[]","name": "accounts","type": "address[]"},
					{
						"components": [
							{"internalType": "bool","name": "b","type": "bool"},
							{"internalType": "bytes","name": "bs","type": "bytes"},
							{"internalType": "bytes24","name": "fbs","type": "bytes24"}
						],
						"internalType": "struct TupleTest.tuple2[]","name": "tuples","type": "tuple[]"
					}
				],
				"internalType": "struct TupleTest.tuple1","name": "t","type": "tuple"
			}
		],
		"stateMutability": "view",
		"type": "function"
	}
]`
)

/******************************************************
*	test solidity code


// SPDX-License-Identifier: MIT
// Compatible with OpenZeppelin Contracts ^5.0.0
pragma solidity ^0.8.20;

contract TupleTest {
    constructor() {}

    struct tuple1 {
        uint256     x256;
        uint24      y24;
        int256      z256;
        string      str;
        address     account;
        address[]   accounts;
        tuple2[]    tuples;
    }
    struct tuple2 {
        bool        b;
        bytes       bs;
        bytes24     fbs;
    }

    uint256 public sum;
    tuple1  public tupleTest;

    function check() public view returns (tuple1 memory) {
        return tupleTest;
    }

    function fnty_1(uint256 x) public pure returns (uint256) {
        return x;
    }
    function fnty_2(uint256 x) public pure returns (uint256) {
        return x+10;
    }
    function argCallback(bytes32 x, uint256 y, function (uint256) external returns (uint256) callback) public {
        sum += uint256(x) + callback(y);
    }
    function argTuple(tuple1 calldata t) public {
        sum += t.x256;
        tupleTest = t;
    }
}
*/
