package komblobulate

import (
    "crypto/cipher"
    "errors"
    "io"
    )

const (
    // The amount by which plaintext expanded into
    // ciphertext
    ExpansionAmount = 16
    )

type AeadReaderWorker struct {
    Config *AeadConfig
    Aead cipher.AEAD

    CipherText io.Reader
    CipherTextLength int

    Nonce, Sealed, Chunk, Ad []byte

    // This tracks how much ciphertext we've read
    TextRead int
}

func (a *AeadReaderWorker) ExpectedLength() int {
    // A ciphertext chunk is our chunk size, plus the nonce size,
    // plus the expansion amount:
    plainChunkSize := int(a.Config.ChunkSize)
    cipherChunkSize := plainChunkSize + PreludeSize + ExpansionAmount + NonceSize

    // Work out how many of these we would theoretically
    // be reading.
    cipherChunkCount := a.CipherTextLength / cipherChunkSize
    cipherChunkLeftOver := a.CipherTextLength % cipherChunkSize

    if cipherChunkLeftOver == 0 {
        return cipherChunkCount * plainChunkSize
    } else {
        // The leftover will have the full nonce and expansion
        // amounts, but only whatever else fits within:
        return cipherChunkCount * plainChunkSize + cipherChunkLeftOver - PreludeSize - ExpansionAmount - NonceSize
    }
}

func (a *AeadReaderWorker) Ready(putChunk func([]byte) error) (err error) {

    // Read the nonce:
    var n int
    n, err = ReadAllOf(a.CipherText, a.Nonce, 0)
    if err != nil {
        return
    }

    a.TextRead += n
    if n != NonceSize {
        err = errors.New("Truncated nonce")
        return
    }

    // Work out how much ciphertext there is left,
    // and only read a truncated section 
    textLeft := a.CipherTextLength - a.TextRead
    if textLeft <= 0 {
        err = io.EOF
        return
    } else if textLeft < len(a.Sealed) {
        a.Sealed = a.Sealed[:textLeft]
    }
        
    n, err = ReadAllOf(a.CipherText, a.Sealed, 0)
    if err != nil {
        return
    }

    a.TextRead += n

    a.Chunk = a.Chunk[:0]
    a.Chunk, err = a.Aead.Open(a.Chunk, a.Nonce, a.Sealed, a.Ad)
    if err != nil {
        return
    }

    // When writing to the buf, skip the prelude,
    // which was just there out of an abundance of
    // crypto caution
    err = putChunk(a.Chunk[PreludeSize:])
    return
}

func NewAeadReader(config *AeadConfig, aead cipher.AEAD, outer io.Reader, outerLength int) *WorkerReader {
    worker := &AeadReaderWorker{
        Config: config,
        Aead: aead,
        CipherText: outer,
        CipherTextLength: outerLength,
        Nonce: make([]byte, NonceSize),
        Sealed: make([]byte, config.ChunkSize + PreludeSize + ExpansionAmount),
        Chunk: make([]byte, config.ChunkSize + PreludeSize),
        TextRead: 0,
    }

    return NewWorkerReader(worker)
}

