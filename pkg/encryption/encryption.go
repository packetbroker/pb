// Copyright © 2018 The Packet Broker Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package encryption implements the encryption and decryption mechanism using AES-128 CBC with
// PKCS#7 padding, and verification with HMAC using SHA-256.
package encryption

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
)

// Encrypt encrypts the specified buffer with the key.
func Encrypt(buf, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// PKCS#7 padding.
	padding := byte(aes.BlockSize - len(buf)%aes.BlockSize)
	if padding == 0 {
		padding = aes.BlockSize
	}
	padded := append(buf, bytes.Repeat([]byte{padding}, int(padding))...)

	// Using a random IV.
	s := make([]byte, aes.BlockSize+len(padded))
	iv := s[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(s[aes.BlockSize:], padded)
	return s, nil
}

// Decrypt decrypts the specified data with the key.
func Decrypt(data, key []byte) ([]byte, error) {
	if len(data) < aes.BlockSize {
		return nil, errors.New("data too short")
	}
	if len(data)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("data is not a multiple of the block size %v", aes.BlockSize)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	s := make([]byte, len(data))
	copy(s, data)
	iv := s[:aes.BlockSize]
	s = s[aes.BlockSize:]

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(s, s)

	// Check PKCS#7 padding; last padding bytes should have value padding.
	padding := s[len(s)-1]
	for _, b := range s[len(s)-int(padding):] {
		if b != padding {
			return nil, errors.New("invalid padding")
		}
	}

	return s[:len(s)-int(padding)], nil
}
