/* AES128-GCM AEAD encryption.
 * No guarantee of actual security :P
 */

package komblobulate

import (
    "bytes"
    "encoding/binary"
    "io"
    )

type AeadConfig struct {
    StartingNonce []byte
}

func (c *AeadConfig) WriteConfig(writer io.Writer) (err error) {
    buf := bytes.NewBuffer(make([]byte, 0, ConfigSize))
    err = binary.Write(buf, binary.LittleEndian, c)
    if err != nil {
        return err
    }

    encoded := buf.Bytes()
    if len(encoded) > ConfigSize {
        panic("Bad aead config size")
    }

    _, err = writer.Write(append(encoded, make([]byte, ConfigSize - len(encoded))...))
    return err
}

func (c *AeadConfig) NewReader(outer io.Reader, p ...interface{}) (inner io.Reader, err error) {
    // TODO TODO
    return nil, nil
}

func (c *AeadConfig) NewWriter(outer io.Writer, p ...interface{}) (inner io.WriteCloser, err error) {
    // TODO TODO
    return nil, nil
}

func NewAeadConfig() (*AeadConfig, error) {
    // TODO Generate the starting nonce, etc.
    return nil, nil
}

func ReadAeadConfig(reader io.Reader) (*AeadConfig, error) {
    config := new(AeadConfig)
    err := binary.Read(reader, binary.LittleEndian, config)
    return config, err
}

