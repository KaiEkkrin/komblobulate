package komblobulate

import (
    "bytes"
    "crypto/cipher"
    "crypto/rand"
    "errors"
    "io"
    )

type AeadWriter struct {
    Config *AeadConfig
    Aead cipher.AEAD

    CipherText io.Writer

    // We'll run a goroutine to deal with sealing up
    // the bytes fed into us, and message it with
    // whichever of our two buffers is ready.  -1
    // tells it to stop.
    Bufs [2]bytes.Buffer
    CurrentInputBuf int
    Ready chan int
    Finished chan error
}

// The encryption goroutine:
func (a *AeadWriter) Encrypt() {
    var err error
    defer func() {
        a.Finished <- err
    }()

    nonce := make([]byte, 12)
    chunk := make([]byte, a.Config.ChunkSize)
    var sealed, ad []byte

    more := true
    for more {
        readyBuf := <- a.Ready
        if readyBuf == -1 {
            more = false
        } else {

            // Generate a new nonce:
            n, err := rand.Read(nonce)
            if err == nil && n != len(nonce) {
                err = errors.New("Read bad nonce length")
            }

            if err != nil {
                return
            }

            // Write it first of all:
            n, err = a.CipherText.Write(nonce)
            if err != nil {
                return
            }

            // Pull as much ciphertext as I can, up to a max
            // of the chunk size:
            n, err = a.Bufs[readyBuf].Read(chunk)
            if err != nil {
                return
            }

            // Seal it up.
            // TODO Do I want nonzero additional data?
            sealed = sealed[:0]
            sealed = a.Aead.Seal(sealed, nonce, chunk[:n], ad)

            // Write all that to the output:
            n, err = WriteAllOf(a.CipherText, sealed, 0)
            if err != nil {
                return
            }

            // I finished with that buffer:
            a.Bufs[readyBuf].Reset()
        }
    }
}

func (a *AeadWriter) Write(p []byte) (n int, err error) {
    // TODO Should I enforce thread safety by having this
    // in another goroutine?
    for len(p) > 0 {
        // Fill the current input buf up to the chunk size:
        lenToWrite := int(a.Config.ChunkSize) - a.Bufs[a.CurrentInputBuf].Len()
        if lenToWrite > len(p) {
            lenToWrite = len(p)
        }

        toWrite := p[:lenToWrite]
        p = p[lenToWrite:]

        var written int
        written, err = a.Bufs[a.CurrentInputBuf].Write(toWrite)
        n += written
        if err != nil {
            return
        }

        if written < lenToWrite {
            p = append(toWrite[written:], p...)
        }

        if a.Bufs[a.CurrentInputBuf].Len() == int(a.Config.ChunkSize) {
            // Write this one and roll over to the other
            // input buffer:
            a.Ready <- a.CurrentInputBuf
            a.CurrentInputBuf = 1 - a.CurrentInputBuf 
        }
    }

    return
}

func (a *AeadWriter) Close() error {
    // Write whatever's left over in the current input buf:
    if a.Bufs[a.CurrentInputBuf].Len() > 0 {
        a.Ready <- a.CurrentInputBuf
        a.CurrentInputBuf = 1 - a.CurrentInputBuf
    }

    // Tell that goroutine to finish:
    a.Ready <- -1
    return <- a.Finished
}

func NewAeadWriter(config *AeadConfig, aead cipher.AEAD, outer io.Writer) *AeadWriter {
    writer := &AeadWriter{
        Config: config,
        Aead: aead,
        CipherText: outer,
        CurrentInputBuf: 0,
        Ready: make(chan int),
        Finished: make(chan error, 1),
    }

    go writer.Encrypt()
    return writer
}

