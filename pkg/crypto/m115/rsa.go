package m115

import (
	"bytes"
	"crypto/rand"
	"io"
	"math/big"
)

var (
	_N, _ = big.NewInt(0).SetString(
		"8686980c0f5a24c4b9d43020cd2c22703ff3f450756529058b1cf88f09b86021"+
			"36477198a6e2683149659bd122c33592fdb5ad47944ad1ea4d36c6b172aad633"+
			"8c3bb6ac6227502d010993ac967d1aef00f0c8e038de2e4d3bc2ec368af2e9f1"+
			"0a6f1eda4f7262f136420c07c331b871bf139f74f3010e3c4fe57df3afb71683", 16)
	_E, _ = big.NewInt(0).SetString("10001", 16)

	_KeyLength = _N.BitLen() / 8
)

func rsaEncrypt(input []byte) []byte {
	buf := &bytes.Buffer{}
	for remainSize := len(input); remainSize > 0; {
		sliceSize := _KeyLength - 11
		if sliceSize > remainSize {
			sliceSize = remainSize
		}
		rsaEncryptSlice(input[:sliceSize], buf)

		input = input[sliceSize:]
		remainSize -= sliceSize
	}
	return buf.Bytes()
}

func rsaEncryptSlice(input []byte, w io.Writer) {
	// Padding
	padSize := _KeyLength - len(input) - 3
	padData := make([]byte, padSize)
	_, _ = rand.Read(padData)
	// Prepare message
	buf := make([]byte, _KeyLength)
	buf[0], buf[1] = 0, 2
	for i, b := range padData {
		buf[2+i] = b%0xff + 0x01
	}
	buf[padSize+2] = 0
	copy(buf[padSize+3:], input)
	msg := big.NewInt(0).SetBytes(buf)
	// RSA Encrypt
	ret := big.NewInt(0).Exp(msg, _E, _N).Bytes()
	// Fill zeros at beginning
	if fillSize := _KeyLength - len(ret); fillSize > 0 {
		zeros := make([]byte, fillSize)
		_, _ = w.Write(zeros)
	}
	_, _ = w.Write(ret)
}

func rsaDecrypt(input []byte) []byte {
	buf := &bytes.Buffer{}
	for remainSize := len(input); remainSize > 0; {
		sliceSize := _KeyLength
		if sliceSize > remainSize {
			sliceSize = remainSize
		}
		rsaDecryptSlice(input[:sliceSize], buf)

		input = input[sliceSize:]
		remainSize -= sliceSize
	}
	return buf.Bytes()
}

func rsaDecryptSlice(input []byte, w io.Writer) {
	// RSA Decrypt
	msg := big.NewInt(0).SetBytes(input)
	ret := big.NewInt(0).Exp(msg, _E, _N).Bytes()
	// Un-padding
	for i, b := range ret {
		// Find the beginning of plaintext
		if b == 0 && i != 0 {
			_, _ = w.Write(ret[i+1:])
			break
		}
	}
}
