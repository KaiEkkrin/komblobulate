/* Reed-Solomon resist wrapping. */

package komblobulate

import (
    "bytes"
    "encoding/binary"
    "io"
    )

type RsConfig struct {
    // The byte size of each piece, minus the CRC
    // (which is 4 bytes).
    DataPieceSize int

    // The number of data pieces in each separate
    // RS matrix.
    DataPieceCount int

    // The number of parity pieces in each separate
    // RS matrix.
    ParityPieceCount int
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

    _, err = WriteAllOf(writer, append(encoded, make([]byte, ConfigSize - len(encoded))...), 0)
    return err
}

func (c *RsConfig) NewReader(outer io.Reader, outerLength int, p ...interface{}) (inner io.Reader, innerLength int, err error) {
    // TODO TODO
    return nil, 0, nil
}

func (c *RsConfig) NewWriter(outer io.Writer, p ...interface{}) (inner io.WriteCloser, err error) {
    // TODO TODO
    return nil, nil
}

func ReadRsConfig(reader io.Reader) (*RsConfig, error) {
    config := new(RsConfig)
    err := binary.Read(reader, binary.LittleEndian, config)
    return config, err
}

