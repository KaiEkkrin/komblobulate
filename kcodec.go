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
    
    // Wraps a reader in a reader that decodes this kcodec.
    // Accepts codec-specific arbitrary parameters (e.g.
    // an encryption key for aead).
    NewReader(io.Reader, ...interface{}) (io.Reader, error)

    // Wraps a writer in a writer that decodes this kcodec.
    NewWriter(io.Writer, ...interface{}) (io.WriteCloser, error)
}

