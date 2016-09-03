/* Komblobulate is a streaming bytes container type (providing a reader
 * and a writer) that applies encryption and error resistance to its
 * input.
 * (Which should probably be compressed, because komblobulate is going
 * to make it incompressible.)
*/

package komblobulate

import (
    "io"
    )

const (
    ResistType_None = 0
    ResistType_Rs = 1

    CipherType_None = 0
    CipherType_Aead = 1
    )

type Config struct {
    ResistType int
    CipherType int
}

// Given a reader of a kblobbed output, creates a reader of the
// unblobbed contents.  The kblob itself will contain its
// configuration.
func NewReader(kblob io.ReadSeeker) (unblob io.Reader, err error) {
    // TODO.
    return nil, nil
}

// Given a writer of where the user wants the kblobbed output to
// go and a configuration, creates a writer for unblobbed contents.
// resistConfig should be:
// - nil if ResistType is ResistType_None
// - *RsConfig if ResistType is ResistType_Rs
// cipherConfig should be:
// - nil if CipherType is CipherType_None
// - *AeadConfig if CipherType is CipherType_Aead
// The variable parameters are passed to the cipher config.
// AeadConfig, for example, will expect a single passphrase
// (which it will munge into an AES key).
func NewWriter(kblob io.Writer, config Config, resistConfig interface{}, cipherConfig interface{}, p ...interface{}) (unblob io.WriteCloser, err error) {
    // TODO.
    return nil, nil
}

