package ec115

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"math/big"

	"github.com/aead/ecdh"
	"github.com/andreburgaud/crypt2go/ecb"
	"github.com/andreburgaud/crypt2go/padding"
	"github.com/pierrec/lz4/v4"
)

var remotePubKey = []byte{
	0x57, 0xA2, 0x92, 0x57, 0xCD, 0x23, 0x20, 0xE5,
	0xD6, 0xD1, 0x43, 0x32, 0x2F, 0xA4, 0xBB, 0x8A,
	0x3C, 0xF9, 0xD3, 0xCC, 0x62, 0x3E, 0xF5, 0xED,
	0xAC, 0x62, 0xB7, 0x67, 0x8A, 0x89, 0xC9, 0x1A,
	0x83, 0xBA, 0x80, 0x0D, 0x61, 0x29, 0xF5, 0x22,
	0xD0, 0x34, 0xC8, 0x95, 0xDD, 0x24, 0x65, 0x24,
	0x3A, 0xDD, 0xC2, 0x50, 0x95, 0x3B, 0xEE, 0xBA,
}

const (
	p224BaseLen = 28
	crcSalt     = "^j>WD3Kr?J2gLFjD4W2y@"
)

// 利用key进行异或操作
func xor(src, key []byte) []byte {
	secret := make([]byte, 0, len(src))
	pad := len(src) % 4
	if pad > 0 {
		for i := 0; i < pad; i++ {
			secret = append(secret, src[i]^key[i])
		}
		src = src[pad:]
	}
	keyLen := len(key)
	num := 0
	for _, s := range src {
		if num >= keyLen {
			num = num % keyLen
		}
		secret = append(secret, s^key[num])
		num++
	}

	return secret
}

// EcdhCipher ECDH加密解密信息
type EcdhCipher struct {
	key    []byte
	iv     []byte
	pubKey []byte
}

// NewEcdhCipher 新建EcdhCipher
func NewEcdhCipher() (*EcdhCipher, error) {
	x := big.NewInt(0).SetBytes(remotePubKey[:p224BaseLen])
	y := big.NewInt(0).SetBytes(remotePubKey[p224BaseLen:])
	remotePublic := ecdh.Point{X: x, Y: y}

	p224 := ecdh.Generic(elliptic.P224())
	private, public, err := p224.GenerateKey(rand.Reader)

	buf := make([]byte, p224BaseLen)
	switch p := public.(type) {
	case ecdh.Point:
		p.X.FillBytes(buf)
		if big.NewInt(0).And(p.Y, big.NewInt(1)).Cmp(big.NewInt(1)) == 0 {
			buf = append([]byte{p224BaseLen + 1, 0x03}, buf...)
		} else {
			buf = append([]byte{p224BaseLen + 1, 0x02}, buf...)
		}
	default:
		return nil, fmt.Errorf("错误的public key类型")
	}
	if err != nil {
		return nil, err
	}

	secret := p224.ComputeSecret(private, remotePublic)

	cipher := new(EcdhCipher)
	cipher.key = secret[:aes.BlockSize]
	cipher.iv = secret[len(secret)-aes.BlockSize:]
	cipher.pubKey = buf
	return cipher, nil
}

// Encrypt 加密
func (c *EcdhCipher) Encrypt(plainText []byte) ([]byte, error) {
	pad := padding.NewPkcs7Padding(aes.BlockSize)
	data, err := pad.Pad(plainText)
	if err != nil {
		return nil, err
	}

	cipherText := make([]byte, 0, len(data))
	var xorKey []byte
	xorKey = append(xorKey, c.iv...)
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}
	mode := ecb.NewECBEncrypter(block)
	tmp := make([]byte, 0, aes.BlockSize)

	for i, b := range data {
		tmp = append(tmp, b^xorKey[i%aes.BlockSize])
		if i%aes.BlockSize == aes.BlockSize-1 {
			mode.CryptBlocks(xorKey, tmp)
			cipherText = append(cipherText, xorKey...)
			tmp = make([]byte, 0, aes.BlockSize)
		}
	}

	return cipherText, nil
}

// Decrypt 解密
func (c *EcdhCipher) Decrypt(cipherText []byte) (text []byte, e error) {
	defer func() {
		if err := recover(); err != nil {
			e = fmt.Errorf("%v", err)
		}
	}()

	cipherText = cipherText[0 : len(cipherText)-len(cipherText)%aes.BlockSize]

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	lz4Block := make([]byte, len(cipherText))
	mode := cipher.NewCBCDecrypter(block, c.iv)
	mode.CryptBlocks(lz4Block, cipherText)

	length := int(lz4Block[0]) + int(lz4Block[1])<<8
	text = make([]byte, 0x2000)
	l, err := lz4.UncompressBlock(lz4Block[2:length+2], text)
	if err != nil {
		return nil, err
	}

	return text[:l], nil
}

// EncodeToken 加密token
func (c *EcdhCipher) EncodeToken(timestamp int64) (string, error) {
	random, err := rand.Int(rand.Reader, big.NewInt(256))
	if err != nil {
		return "", err
	}
	r1 := byte(random.Uint64())
	random, err = rand.Int(rand.Reader, big.NewInt(256))
	if err != nil {
		return "", err
	}
	r2 := byte(random.Uint64())
	tmp := make([]byte, 0, 48)

	time := make([]byte, 4)
	binary.BigEndian.PutUint32(time, uint32(timestamp))

	for i := 0; i < 15; i++ {
		tmp = append(tmp, c.pubKey[i]^r1)
	}
	tmp = append(tmp, []byte{r1, 0x73 ^ r1}...)
	for i := 0; i < 3; i++ {
		tmp = append(tmp, r1)
	}
	for i := 0; i < 4; i++ {
		tmp = append(tmp, r1^time[3-i])
	}
	for i := 15; i < len(c.pubKey); i++ {
		tmp = append(tmp, c.pubKey[i]^r2)
	}
	tmp = append(tmp, []byte{r2, 0x01 ^ r2}...)
	for i := 0; i < 3; i++ {
		tmp = append(tmp, r2)
	}

	crc := make([]byte, 4)
	binary.BigEndian.PutUint32(crc, crc32.ChecksumIEEE(append([]byte(crcSalt), tmp...)))

	for i := 0; i < 4; i++ {
		tmp = append(tmp, crc[3-i])
	}

	return base64.StdEncoding.EncodeToString(tmp), nil
}
