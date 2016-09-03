/* Komblobulate is a streaming bytes container type (providing a reader
 * and a writer) that applies encryption and error resistance to its
 * input.
 * (Which should probably be compressed, because komblobulate is going
 * to make it incompressible.)
*/

package komblobulate

import (
    "io"
    )

const (
    ResistType_None = 0
    ResistType_Rs = 1

    CipherType_None = 0
    CipherType_Aead = 1
    )

// Given a reader of a kblobbed output, creates a reader of the
// unblobbed contents.  The kblob itself will contain its
// configuration.
func NewReader(kblob io.ReadSeeker) (unblob io.Reader, err error) {
    // TODO.
    return nil, nil
}

// Given a writer of where the user wants the kblobbed output to
// go and a configuration, creates a writer for unblobbed contents.
// The variable parameters should contain the relevant parameters
// to create the resist codec, followed by those to create the
// cipher codec.  For example:
// - a none config wants none
// - rs wants three integers DataPieceSize, DataPieceCount,
// ParityPieceCount
// - aead wants a passphrase.
func NewWriter(kblob io.Writer, resistType int, cipherType int, p ...interface{}) (unblob io.WriteCloser, err error) {

    pIdx := 0

    var resist, cipher KCodec
    switch resistType {
    case ResistType_None:
        resist = &NullConfig{}

    case ResistType_Rs:
        dataPieceSize, ok := p[pIdx].(int)
        pIdx++
        if !ok {
            panic("Bad Rs parameter")
        }

        dataPieceCount, ok := p[pIdx].(int)
        pIdx++
        if !ok {
            panic("Bad Rs parameter")
        }

        parityPieceCount, ok := p[pIdx].(int)
        pIdx++
        if !ok {
            panic("Bad Rs parameter")
        }

        resist = &RsConfig{dataPieceSize, dataPieceCount, parityPieceCount}

    default:
        panic("Bad resist type")
    }

    switch cipherType {
    case CipherType_None:
        cipher = &NullConfig{}

    case CipherType_Aead:
        cipher, err = NewAeadConfig()
        if err != nil {
            return nil, err
        }

    default:
        panic("Bad cipher type")
    }

    config := &Config{resistType, cipherType}

    // Write the whole config twice at the start:
    err = config.WriteConfig(kblob, resist, cipher)
    if err != nil {
        return
    }

    err = config.WriteConfig(kblob, resist, cipher)
    if err != nil {
        return
    }

    // Create the inner writers:
    resistWriter, err := resist.NewWriter(kblob)
    if err != nil {
        return
    }

    cipherWriter, err := cipher.NewWriter(resistWriter, p[pIdx:]...)
    if err != nil {
        return
    }

    unblob = &KblobWriter{config, resist, cipher, kblob, resistWriter, cipherWriter}
    return
}

