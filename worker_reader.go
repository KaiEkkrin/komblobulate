/* The counterpart to worker_writer: a generic way of
 * building a streaming reader system with a goroutine
 * to process chunks of data.
*/

package komblobulate

import (
    "bytes"
    "io"
    )

type ReaderWorker interface {
    // Called whenever we're ready to receive a chunk,
    // with a function that will write the chunk
    // into the read buffer.
    Ready(func([]byte) error) error
}

type WorkerReader struct {
    Worker ReaderWorker

    Bufs [2]bytes.Buffer
    CurrentInputBuf int
    Ready chan int
}

func (w *WorkerReader) Decode(bufIdx int) {
    defer func() {
        // Forever indicate we've finished on request:
        for {
            w.Ready <- -1
        }
    }()

    for {
        workErr := w.Worker.Ready(func(chunk []byte) (err error) {
            // Send along these data
            _, err = w.Bufs[bufIdx].Write(chunk)
            if err != nil {
                return err
            }

            w.Ready <- bufIdx

            // Switch to the other buffer for the next one
            bufIdx = 1 - bufIdx
            w.Bufs[bufIdx].Reset()

            return
        })

        // TODO I have no exit route for this error right now
        if workErr != nil {
            panic(workErr.Error())
        }
    }
}

func (w *WorkerReader) Read(p []byte) (n int, err error) {
    // If we've run out of data, wait for more:
    if w.Bufs[w.CurrentInputBuf].Len() == 0 {
        w.CurrentInputBuf = <- w.Ready
    }
        
    // The end-of-file condition:
    if w.CurrentInputBuf == -1 {
        return 0, io.EOF
    }

    // Read however much is in the buffer:
    n, err = w.Bufs[w.CurrentInputBuf].Read(p)
    return
}

func NewWorkerReader(worker ReaderWorker) *WorkerReader {
    reader := &WorkerReader{
        Worker: worker,
        CurrentInputBuf: 0,
        Ready: make(chan int),
    }

    go reader.Decode(1)
    return reader
}

