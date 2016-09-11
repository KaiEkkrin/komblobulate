/* A generic way of building a streaming writer system
 * with a goroutine to process chunks of data
 * and a buffer pair for accepting it and feeding
 * chunk sized pieces to the goroutine
*/

package komblobulate

import (
    "bytes"
    )

type WriterWorker interface {
    // The chunk size to send to the worker;
    // this should be a constant value.
    ChunkSize() int

    // Called whenever a chunk is ready, with a
    // function that reads the chunk into a slice and
    // returns how many bytes were read.
    Ready(func([]byte) (int, error)) error

    // Called at the end of the encode to finish up:
    Close() error
}

type WorkerWriter struct {
    Worker WriterWorker

    Bufs [2]bytes.Buffer
    CurrentInputBuf int
    Ready chan int
    Finished chan error
}

func (w *WorkerWriter) Encode() {
    var err error
    defer func() {
        w.Finished <- err
    }()

    more := true
    for more {
        readyBuf := <- w.Ready
        if readyBuf == -1 {

            // Finished.
            more = false

        } else {

            // The worker can read the next chunk:
            err = w.Worker.Ready(func(chunk []byte) (int, error) {
                return w.Bufs[readyBuf].Read(chunk)
            })

            if err != nil {
                return
            }

            // I finished with that buffer:
            w.Bufs[readyBuf].Reset()
        }
    }

    err = w.Worker.Close()
}

func (w *WorkerWriter) Write(p []byte) (n int, err error) {
    // TODO Should I enforce thread safety by having this
    // in another goroutine?
    chunkSize := w.Worker.ChunkSize()
    for len(p) > 0 {
        // Fill the current input buf up to the chunk size:
        lenToWrite := chunkSize - w.Bufs[w.CurrentInputBuf].Len()
        if lenToWrite > len(p) {
            lenToWrite = len(p)
        }

        toWrite := p[:lenToWrite]
        p = p[lenToWrite:]

        var written int
        written, err = w.Bufs[w.CurrentInputBuf].Write(toWrite)
        n += written
        if err != nil {
            return
        }

        if written < lenToWrite {
            p = append(toWrite[written:], p...)
        }

        if w.Bufs[w.CurrentInputBuf].Len() == chunkSize {
            // Write this one and roll over to the other
            // input buffer:
            select {
            case w.Ready <- w.CurrentInputBuf:
                w.CurrentInputBuf = 1 - w.CurrentInputBuf 

            case err = <- w.Finished:
                return
            }
        }
    }

    return
}

func (w *WorkerWriter) Close() (err error) {
    // Write whatever's left over in the current input buf:
    if w.Bufs[w.CurrentInputBuf].Len() > 0 {
        select {
        case w.Ready <- w.CurrentInputBuf:
            w.CurrentInputBuf = 1 - w.CurrentInputBuf

        case err = <- w.Finished:
            return
        }
    }

    // Tell that goroutine to finish:
    w.Ready <- -1
    return <- w.Finished
}

func (w *WorkerWriter) UpdateConfig(config KCodec) {
    w.UpdateConfig(config)
}

func NewWorkerWriter(worker WriterWorker) *WorkerWriter {
    writer := &WorkerWriter{
        Worker: worker,
        CurrentInputBuf: 0,
        Ready: make(chan int),
        Finished: make(chan error, 1),
    }

    go writer.Encode()
    return writer
}

