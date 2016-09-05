package komblobulate

import (
    "bytes"
    "crypto/cipher"
    "errors"
    "io"
    )

const (
    // The amount by which plaintext expanded into
    // ciphertext
    ExpansionAmount = 16
    )

type AeadReader struct {
    Config *AeadConfig
    Aead cipher.AEAD

    CipherText io.Reader
    CipherTextLength int

    // This works similarly to the writer,
    // but in this case, it's the goroutine that will
    // send the buf index to Ready when a buf is
    // ready.  CurrentInputBuf is the current *plaintext*
    // input buf that Read() should be reading from.
    // We don't have a close equivalent, but we send -1
    // when finished.
    Bufs [2]bytes.Buffer
    CurrentInputBuf int
    Ready chan int
}

// The decryption goroutine.
// bufIdx should track the one that Read() *isn't*
// currently reading from.
func (a *AeadReader) Decrypt(bufIdx int) {
    var err error
    defer func() {
        // I have no exit route for this!
        if err != nil && err != io.EOF {
            panic(err.Error())
        }

        // Forever indicate we've finished on request:
        for {
            a.Ready <- -1
        }
    }()

    // Keep track of how much ciphertext we've read:
    textRead := 0

    nonce := make([]byte, NonceSize)
    sealed := make([]byte, a.Config.ChunkSize + PreludeSize + ExpansionAmount)
    chunk := make([]byte, a.Config.ChunkSize + PreludeSize)
    var ad []byte

    for {

        // Read the nonce:
        var n int
        n, err = ReadAllOf(a.CipherText, nonce, 0)
        if err != nil {
            return
        }

        textRead += n
        if n != NonceSize {
            err = errors.New("Truncated nonce")
            return
        }

        // Work out how much ciphertext there is left,
        // and only read a truncated section 
        textLeft := a.CipherTextLength - textRead
        if textLeft <= 0 {
            err = io.EOF
            return
        } else if textLeft < len(sealed) {
            sealed = sealed[:textLeft]
        }
            
        n, err = ReadAllOf(a.CipherText, sealed, 0)
        if err != nil {
            return
        }

        textRead += n

        chunk = chunk[:0]
        chunk, err = a.Aead.Open(chunk, nonce, sealed, ad)
        if err != nil {
            return
        }

        // When writing to the buf, skip the prelude,
        // which was just there out of an abundance of
        // crypto caution
        a.Bufs[bufIdx].Reset()
        n, err = a.Bufs[bufIdx].Write(chunk[PreludeSize:])
        if err != nil {
            return
        }

        a.Ready <- bufIdx
        bufIdx = 1 - bufIdx
    }
}

func (a *AeadReader) ExpectedLength() int {
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
        return cipherChunkCount * plainChunkSize + cipherChunkLeftOver - PreludeSize - ExpansionAmount - 12
    }
}

func (a *AeadReader) Read(p []byte) (n int, err error) {
    // If we've run out of data, wait for more:
    if a.Bufs[a.CurrentInputBuf].Len() == 0 {
        a.CurrentInputBuf = <- a.Ready
    }
        
    // The end-of-file condition:
    if a.CurrentInputBuf == -1 {
        return 0, io.EOF
    }

    // Read however much is in the buffer:
    return a.Bufs[a.CurrentInputBuf].Read(p)
}

func NewAeadReader(config *AeadConfig, aead cipher.AEAD, outer io.Reader, outerLength int) *AeadReader {
    reader := &AeadReader{
        Config: config,
        Aead: aead,
        CipherText: outer,
        CipherTextLength: outerLength,
        CurrentInputBuf: 0,
        Ready: make(chan int),
    }

    go reader.Decrypt(1)
    return reader
}

