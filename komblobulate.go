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
	ResistType_Rs   = byte(1)

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
func NewReader(kblob io.ReadSeeker, params KCodecParams) (unblob io.Reader, err error) {

	// Work out how big this blob is:
	var bloblen int64
	bloblen, err = kblob.Seek(0, 2)
	if err != nil {
		return nil, err
	}

	// The config is stored in three places -- twice at
	// the beginning and once at the end.  Read out
	// all three, ignoring errors so long as we manage
	// to get agreement:
	var configBlocks [3]interface{}
	var resistBlocks [3]interface{}
	var cipherBlocks [3]interface{}

	configBlocks[0], resistBlocks[0], cipherBlocks[0], err = ReadConfig(kblob, 0)
	configBlocks[1], resistBlocks[1], cipherBlocks[1], err = ReadConfig(kblob, 3*ConfigSize)
	configBlocks[2], resistBlocks[2], cipherBlocks[2], err = ReadConfig(kblob, bloblen-3*ConfigSize)

	config, ok := findAgreement(configBlocks, func(a interface{}, b interface{}) bool {
		return a.(*Config).ConfigEquals(b.(*Config))
	}).(*Config)

	if !ok || config == nil {
		if err == nil {
			err = errors.New("No config agreement")
		}
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

	unResist, unResistLength, err := resist.NewReader(unConfig, bloblen-int64(9*ConfigSize), params)
	if err != nil {
		return
	}

	unblob, _, err = cipher.NewReader(unResist, unResistLength, params)
	return
}

// Given a writer of where the user wants the kblobbed output to
// go and a configuration, creates a writer for unblobbed contents.
func NewWriter(kblob io.WriteSeeker, resistType byte, cipherType byte, params KCodecParams) (unblob io.WriteCloser, err error) {

	var resist, cipher KCodec
	switch resistType {
	case ResistType_None:
		resist = &NullConfig{}

	case ResistType_Rs:
		dataPieceSize, dataPieceCount, parityPieceCount := params.GetRsParams()
		resist = &RsConfig{int32(dataPieceSize), int8(dataPieceCount), int8(parityPieceCount), 0}

	default:
		panic("Bad resist type")
	}

	switch cipherType {
	case CipherType_None:
		cipher = &NullConfig{}

	case CipherType_Aead:
		cipher, err = NewAeadConfig(int64(params.GetAeadChunkSize()))
		if err != nil {
			panic(err.Error())
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
	resistWriter, err := resist.NewWriter(kblob, params)
	if err != nil {
		return
	}

	cipherWriter, err := cipher.NewWriter(resistWriter, params)
	if err != nil {
		return
	}

	unblob = &KblobWriter{config, resist, cipher, kblob, resistWriter, cipherWriter}
	return
}
