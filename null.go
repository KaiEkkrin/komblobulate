/* The null kcodec. */

package komblobulate

import (
    "io"
    )

type NullWriteCloser struct {
    Outer io.Writer
}

func (w *NullWriteCloser) Write(p []byte) (int, error) {
    return w.Outer.Write(p)
}

func (w *NullWriteCloser) Close() error {
    return nil
}

type NullConfig struct {}

func (c *NullConfig) ConfigEquals(other interface{}) bool {
    if other == nil {
        return false
    } else if _, ok := other.(*NullConfig); ok {
        return true
    } else {
        return false
    }
}

func (c *NullConfig) WriteConfig(writer io.Writer) error {
    _, err := WriteAllOf(writer, make([]byte, ConfigSize), 0)
    return err
}

func (c *NullConfig) NewReader(outer io.Reader, outerLength int, p ...interface{}) (io.Reader, int, error) {
    return outer, outerLength, nil
}

func (c *NullConfig) NewWriter(outer io.Writer, p ...interface{}) (io.WriteCloser, error) {
    return &NullWriteCloser{outer}, nil
}

