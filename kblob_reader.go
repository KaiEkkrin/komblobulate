/* Overarching reader stuff.
 * KblobReader is not actually the reader that we return,
 * but rather, the one we can give to the resist reader
 * that will skip the config blocks after they've
 * been validated.
*/

package komblobulate

import (
    "io"
    )

type KblobReader struct {
    Kblob io.ReadSeeker
    Length int64
    Offset int64
}

func (kb *KblobReader) Read(p []byte) (n int, err error) {
    innerEnd := kb.Length - 3 * ConfigSize
    if kb.Offset >= innerEnd {
        return 0, io.EOF
    }

    n, err = kb.Kblob.Read(p)
    kb.Offset += int64(n)

    // Check whether we've read an excess, and truncate
    // it if we have:
    overread := kb.Offset - innerEnd
    if overread > 0 {
        n -= int(overread)
        p = p[:len(p) - int(overread)]
        if err == nil {
            err = io.EOF
        }
    }

    return
}

func NewKblobReader(kblob io.ReadSeeker, bloblen int64) (*KblobReader, error) {
    // Seek past the config blocks:
    offset, err := kblob.Seek(ConfigSize * 6, 0)
    return &KblobReader{kblob, bloblen, offset}, err   
}

