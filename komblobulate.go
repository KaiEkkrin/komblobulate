/* Komblobulate is a streaming bytes container type (providing a reader
 * and a writer) that applies encryption and error resistance to its
 * input.
 * (Which should probably be compressed, because komblobulate is going
 * to make it incompressible.)
*/

package komblobulate

import (
    "errors"
    "io"
    )

const (
    ResistType_None = byte(0)
    ResistType_Rs = byte(1)

    CipherType_None = byte(0)
    CipherType_Aead = byte(1)
    )

// Given three things that might equal each other, finds the
// value that occurs at least twice according to the equality
// function, or nil if they all differ.
func findAgreement(things [3]interface{}, equals func(interface{}, interface{}) bool) interface{} {
    if things[0] != nil {
        if equals(things[0], things[1]) || equals(things[0], things[2]) {
            return things[0]
        }
    }

    if things[1] != nil && equals(things[1], things[2]) {
        return things[1]
    }

    return nil
}

// Given a reader of a kblobbed output, creates a reader of the
// unblobbed contents.  The kblob itself will contain its
// configuration.
// The variable arguments will go to the cipher codec,
// e.g. should contain the passphrase.
func NewReader(kblob io.ReadSeeker, bloblen int64, p ...interface{}) (unblob io.Reader, err error) {

    // The config is stored in three places -- twice at
    // the beginning and once at the end.  Read out
    // all three:
    var configBlocks [3]interface{}
    var resistBlocks [3]interface{}
    var cipherBlocks [3]interface{}

    configBlocks[0], resistBlocks[0], cipherBlocks[0], err = ReadConfig(kblob, 0)
    if err != nil {
        return
    }

    configBlocks[1], resistBlocks[1], cipherBlocks[1], err = ReadConfig(kblob, 3 * ConfigSize)
    if err != nil {
        return
    }

    configBlocks[2], resistBlocks[2], cipherBlocks[2], err = ReadConfig(kblob, bloblen - 3 * ConfigSize)
    if err != nil {
        return
    }

    config, ok := findAgreement(configBlocks, func(a interface{}, b interface{}) bool {
        return a.(*Config).ConfigEquals(b.(*Config))
    }).(*Config)
    if !ok || config == nil {
        err = errors.New("No config agreement")
        return
    }

    kcodecequal := func(a interface{}, b interface{}) bool {
        return a.(KCodec).ConfigEquals(b)
    }

    resist, ok := findAgreement(resistBlocks, kcodecequal).(KCodec)
    if !ok || resist == nil {
        err = errors.New("No resist agreement")
        return
    }

    cipher, ok := findAgreement(cipherBlocks, kcodecequal).(KCodec)
    if !ok || resist == nil {
        err = errors.New("No cipher agreement")
        return
    }

    unConfig, err := NewKblobReader(kblob, bloblen)
    if err != nil {
        return
    }

    unResist, err := resist.NewReader(unConfig)
    if err != nil {
        return
    }

    unblob, err = cipher.NewReader(unResist, p...)
    return
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
func NewWriter(kblob io.Writer, resistType byte, cipherType byte, p ...interface{}) (unblob io.WriteCloser, err error) {

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

