/* AES128-GCM AEAD encryption.
 * No guarantee of actual security :P
 */

package komblobulate

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"golang.org/x/crypto/pbkdf2"
	"io"
)

const (
	NonceSize   = 12
	PreludeSize = 16
)

type AeadConfig struct {
	// This is the size of the chunks of actual plaintext that
	// we encrypt in one go.
	//
	// We actually encipher:
	// - 16 bytes random
	// - plaintext
	//
	// The data that we write to the
	// underlying stream will consist of:
	// - 12 bytes nonce
	// - ciphertext
	ChunkSize int64

	Salt [8]byte
}

func (c *AeadConfig) makeAead(password string) (aead cipher.AEAD, err error) {

	// Derive the key from that password:
	key := pbkdf2.Key([]byte(password), c.Salt[:], 16384, 16, sha256.New)

	// Create an AES encryption with that key:
	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	// And, make an AEAD out of it:
	aead, err = cipher.NewGCM(block)
	return
}

func (c *AeadConfig) ConfigEquals(other interface{}) bool {
	if other == nil {
		return false
	} else if otherAead, ok := other.(*AeadConfig); ok {
		return otherAead.ChunkSize == c.ChunkSize
	} else {
		return false
	}
}

func (c *AeadConfig) WriteConfig(writer io.Writer) (err error) {
	buf := bytes.NewBuffer(make([]byte, 0, ConfigSize))
	err = binary.Write(buf, binary.LittleEndian, c)
	if err != nil {
		return err
	}

	encoded := buf.Bytes()
	if len(encoded) > ConfigSize {
		panic("Bad aead config size")
	}

	_, err = WriteAllOf(writer, append(encoded, make([]byte, ConfigSize-len(encoded))...), 0)
	return err
}

func (c *AeadConfig) NewReader(outer io.Reader, outerLength int64, params KCodecParams) (inner io.Reader, innerLength int64, err error) {

	var aead cipher.AEAD
	aead, err = c.makeAead(params.GetAeadPassword())
	if err != nil {
		return
	}

	reader := NewAeadReader(c, aead, outer, outerLength)

	// The inner length isn't actually used, thus:
	return reader, -1, nil
}

func (c *AeadConfig) NewWriter(outer io.Writer, params KCodecParams) (inner io.WriteCloser, err error) {

	aead, err := c.makeAead(params.GetAeadPassword())
	if err != nil {
		return
	}

	inner = NewAeadWriter(c, aead, outer)
	return
}

func NewAeadConfig(chunkSize int64) (*AeadConfig, error) {
	salt := make([]byte, 8)
	n, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}

	if n != 8 {
		return nil, errors.New("Failed to read salt")
	}

	config := &AeadConfig{ChunkSize: chunkSize}
	copy(config.Salt[:], salt)
	return config, nil
}

func ReadAeadConfig(reader io.Reader) (*AeadConfig, error) {
	config := new(AeadConfig)
	err := binary.Read(reader, binary.LittleEndian, config)
	return config, err
}
