/* The interface for one of the codec components of the
 * kblob system.
*/

package komblobulate

import (
    "io"
    )

const (
    ConfigSize = 32 // bytes
    )

type KCodec interface {
    // Checks for config equality.
    ConfigEquals(other interface{}) bool

    // This *must* write ConfigSize bytes.
    WriteConfig(io.Writer) error
    
    // Wraps a reader in a reader that decodes this kcodec
    // and returns it and the expected length of the decoded
    // data.
    // Accepts codec-specific arbitrary parameters (e.g.
    // an encryption key for aead).  We also require the
    // length of the data available from the reader.
    NewReader(io.Reader, int64, ...interface{}) (io.Reader, int64, error)

    // Wraps a writer in a writer that decodes this kcodec.
    NewWriter(io.Writer, ...interface{}) (io.WriteCloser, error)
}

