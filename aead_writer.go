package komblobulate

import (
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

type AeadWriterWorker struct {
	Config *AeadConfig
	Aead   cipher.AEAD

	CipherText io.Writer

	Nonce, Chunk, Sealed, Ad []byte
}

func FillRandom(buf []byte) {
	n, err := rand.Read(buf)
	if err != nil {
		panic(err.Error())
	}

	if err == nil && n != len(buf) {
		panic(fmt.Sprintf("Tried to read %d random bytes, got %d", len(buf), n))
	}
}

func (a *AeadWriterWorker) ChunkSize() int {
	return int(a.Config.ChunkSize)
}

func (a *AeadWriterWorker) Ready(getPlain func([]byte) (int, error)) (err error) {

	// Generate a new nonce:
	FillRandom(a.Nonce)

	// Write it first of all:
	var n int
	n, err = WriteAllOf(a.CipherText, a.Nonce, 0)
	if err != nil {
		return
	}

	// Make the prelude for the next chunk:
	FillRandom(a.Chunk[:PreludeSize])

	// Read the next chunk of ciphertext:
	n, err = getPlain(a.Chunk[PreludeSize:])
	if err != nil {
		return
	}

	// Seal it up.
	// TODO Do I want nonzero additional data?
	a.Sealed = a.Sealed[:0]
	a.Sealed = a.Aead.Seal(a.Sealed, a.Nonce, a.Chunk[:(n+PreludeSize)], a.Ad)

	// Write all that to the output:
	n, err = WriteAllOf(a.CipherText, a.Sealed, 0)
	return
}

func (a *AeadWriterWorker) Close() error {
	// Nothing to do, we don't buffer internally.
	return nil
}

func NewAeadWriter(config *AeadConfig, aead cipher.AEAD, outer io.Writer) *WorkerWriter {
	worker := &AeadWriterWorker{
		Config:     config,
		Aead:       aead,
		CipherText: outer,
		Nonce:      make([]byte, NonceSize),
		Chunk:      make([]byte, PreludeSize+config.ChunkSize),
	}

	return NewWorkerWriter(worker)
}
