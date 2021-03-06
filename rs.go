/* Reed-Solomon resist wrapping. */

package komblobulate

import (
	"bytes"
	"encoding/binary"
	"io"
)

const (
	CrcSize = 4
)

type RsConfig struct {
	// The byte size of each piece, minus the CRC
	DataPieceSize int32

	// The number of data pieces in each separate
	// RS matrix.
	DataPieceCount int8

	// The number of parity pieces in each separate
	// RS matrix.
	ParityPieceCount int8

	// The total number of data bytes contained inside the
	// reed-solomon encoding.  (The encoded data may have
	// been padded at the end.)
	TotalInnerLength int64
}

func (c *RsConfig) ConfigEquals(other interface{}) bool {
	if other == nil {
		return false
	} else if otherRs, ok := other.(*RsConfig); ok {
		return *otherRs == *c
	} else {
		return false
	}
}

func (c *RsConfig) WriteConfig(writer io.Writer) (err error) {
	buf := bytes.NewBuffer(make([]byte, 0, ConfigSize))
	err = binary.Write(buf, binary.LittleEndian, c)
	if err != nil {
		return err
	}

	encoded := buf.Bytes()
	if len(encoded) > ConfigSize {
		panic("Bad rs config size")
	}

	_, err = WriteAllOf(writer, append(encoded, make([]byte, ConfigSize-len(encoded))...), 0)
	return err
}

func (c *RsConfig) NewReader(outer io.Reader, outerLength int64, params KCodecParams) (inner io.Reader, innerLength int64, err error) {
	return NewRsReader(c, outer), c.TotalInnerLength, nil
}

func (c *RsConfig) NewWriter(outer io.Writer, params KCodecParams) (inner io.WriteCloser, err error) {
	return NewRsWriter(c, outer), nil
}

func ReadRsConfig(reader io.Reader) (*RsConfig, error) {
	config := new(RsConfig)
	err := binary.Read(reader, binary.LittleEndian, config)
	return config, err
}
