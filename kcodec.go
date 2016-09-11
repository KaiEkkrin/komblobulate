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

// How to fetch the optional parameters for the codecs.
// We'll only call these if necessary.  On read,
// only parameters not encoded in the file (i.e.
// GetAeadPassword() ) will be called.
type KCodecParams interface {
    // Returns (data piece size, data piece count, parity piece count).
    GetRsParams() (int, int, int)

    GetAeadChunkSize() int
    GetAeadPassword() string
}

type KCodec interface {
    // Checks for config equality.
    ConfigEquals(other interface{}) bool

    // This *must* write ConfigSize bytes.
    WriteConfig(io.Writer) error
    
    // Wraps a reader in a reader that decodes this kcodec
    // and returns it and the expected length of the decoded
    // data.
    // (reader, length available, parameters).
    NewReader(io.Reader, int64, KCodecParams) (io.Reader, int64, error)

    // Wraps a writer in a writer that decodes this kcodec.
    NewWriter(io.Writer, KCodecParams) (io.WriteCloser, error)
}

