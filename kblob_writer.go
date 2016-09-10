/* Overarching writer stuff. */

package komblobulate

import (
    "io"
    )

type KblobWriter struct {
    Config *Config
    Resist KCodec
    Cipher KCodec

    OuterWriter io.WriteSeeker
    ResistWriter io.WriteCloser
    CipherWriter io.WriteCloser
}

func (kb *KblobWriter) Write(p []byte) (n int, err error) {
    return kb.CipherWriter.Write(p)
}

func (kb *KblobWriter) Close() (err error) {
    err = kb.CipherWriter.Close()
    if err != nil {
        return
    }

    err = kb.ResistWriter.Close()
    if err != nil {
        return
    }

    // Write a final copy of the config at the end:
    err = kb.Config.WriteConfig(kb.OuterWriter, kb.Resist, kb.Cipher)
    if err != nil {
        return
    }

    // Rewrite the copies of the config at the start, since it
    // may have been changed during the write process:
    _, err = kb.OuterWriter.Seek(0, 0)
    if err != nil {
        return
    }

    err = kb.Config.WriteConfig(kb.OuterWriter, kb.Resist, kb.Cipher)
    if err != nil {
        return
    }

    err = kb.Config.WriteConfig(kb.OuterWriter, kb.Resist, kb.Cipher)
    return
}

