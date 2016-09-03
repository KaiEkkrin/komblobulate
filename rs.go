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

    _, err = writer.Write(append(encoded, make([]byte, ConfigSize - len(encoded))...))
    return err
}

func (c *RsConfig) NewReader(outer io.Reader, p ...interface{}) (inner io.Reader, err error) {
    // TODO TODO
    return nil, nil
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

