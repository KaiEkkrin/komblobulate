package komblobulate

import (
    "bytes"
    "encoding/binary"
    "errors"
    "fmt"
    "io"
    )

type Config struct {
    ResistType byte
    CipherType byte
}

func (c *Config) ConfigEquals(other *Config) bool {
    if other == nil {
        return false
    } else {
        return *c == *other
    }
}

func (c *Config) WriteConfig(writer io.Writer, resist KCodec, cipher KCodec) (err error) {

    buf := bytes.NewBuffer(make([]byte, 0, ConfigSize))
    err = binary.Write(buf, binary.LittleEndian, c)
    if err != nil {
        return err
    }

    encoded := buf.Bytes()
    if len(encoded) > ConfigSize {
        panic("Bad base config size")
    }

    _, err = WriteAllOf(writer, append(encoded, make([]byte, ConfigSize - len(encoded))...), 0)

    err = resist.WriteConfig(writer)
    if err != nil {
        return err
    }

    err = cipher.WriteConfig(writer)
    if err != nil {
        return err
    }

    return nil
}

func ReadConfig(reader io.ReadSeeker, offset int64) (config *Config, resist KCodec, cipher KCodec, err error) {
    
    _, err = reader.Seek(offset, 0)
    if err != nil {
        return
    }

    config = new(Config)
    err = binary.Read(reader, binary.LittleEndian, config)
    if err != nil {
        return
    }

    _, err = reader.Seek(offset + ConfigSize, 0)
    if err != nil {
        return
    }

    switch config.ResistType {
    case ResistType_None:
        resist = &NullConfig{}

    case ResistType_Rs:
        resist, err = ReadRsConfig(reader)
        if err != nil {
            return
        }

    default:
        err = errors.New(fmt.Sprintf("Unrecognised resist type %d", config.ResistType))
        return
    }

    _, err = reader.Seek(offset + ConfigSize * 2, 0)
    if err != nil {
        return
    }

    switch config.CipherType {
    case CipherType_None:
        cipher = &NullConfig{}

    case CipherType_Aead:
        cipher, err = ReadAeadConfig(reader)
        if err != nil {
            return
        }

    default:
        err = errors.New(fmt.Sprintf("Unrecognised cipher type %d", config.CipherType))
        return
    }

    return
}

