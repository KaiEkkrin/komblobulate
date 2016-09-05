/* I/O utilities. */

package komblobulate

import (
    "io"
    )

// Reads bytes into buf starting at offset.
// Returns (count read, error), where the output count will
// only ever be less than the input buffer length
// (minus the offset) if we hit an error.
func ReadAllOf(reader io.Reader, buf []byte, offset int) (n int, err error) {
    for offset < len(buf) {
        var thisN int
        thisN, err = reader.Read(buf[offset:])
        offset += thisN
        n += thisN

        if err != nil {
            return
        }
    }

    return
}

// Similarly, writes bytes from buf starting at
// offset, never fewer than the length of buf unless an error
// occurs.
func WriteAllOf(writer io.Writer, buf []byte, offset int) (n int, err error) {
    for offset < len(buf) {
        var thisN int
        thisN, err = writer.Write(buf[offset:])
        offset += thisN
        n += thisN

        if err != nil {
            return
        }
    }

    return
}

