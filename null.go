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

func (c *NullConfig) WriteConfig(writer io.Writer) error {
    _, err := writer.Write(make([]byte, ConfigSize))
    return err
}

func (c *NullConfig) NewReader(outer io.Reader, p ...interface{}) (io.Reader, error) {
    return outer, nil
}

func (c *NullConfig) NewWriter(outer io.Writer, p ...interface{}) (io.WriteCloser, error) {
    return &NullWriteCloser{outer}, nil
}

