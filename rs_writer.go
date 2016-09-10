package komblobulate

// TODO Try using github.com/klauspost/crc32 for
// a large speedup, apparently
import (
    "hash/crc32"
    "io"
    "github.com/klauspost/reedsolomon"
    )

type RsWriterWorker struct {
    Config *RsConfig

    ResistText io.Writer

    Chunk []byte
    Pieces [][]byte
    PieceNum int

    CrcTab *crc32.Table
    Enc reedsolomon.Encoder
}

func (r *RsWriterWorker) ChunkSize() int {
    return r.Config.DataPieceSize
}

func (r *RsWriterWorker) Ready(getPlain func([]byte) (int, error)) (err error) {

    // Read the next chunk:
    var n int
    n, err = getPlain(r.Chunk)
    if err != nil {
        return
    }

    // Update the total length of contained data:
    r.Config.TotalInnerLength += int64(n)

    // Checksum it, and write it and the checksum
    // into the next piece:
    checksum := crc32.Checksum(r.Chunk, r.CrcTab)

    copy(r.Pieces[r.PieceNum], r.Chunk)
    r.Pieces[r.PieceNum][r.Config.DataPieceSize] = byte(checksum & 0xff)
    r.Pieces[r.PieceNum][r.Config.DataPieceSize + 1] = byte((checksum >> 8) & 0xff)
    r.Pieces[r.PieceNum][r.Config.DataPieceSize + 2] = byte((checksum >> 16) & 0xff)
    r.Pieces[r.PieceNum][r.Config.DataPieceSize + 3] = byte((checksum >> 24) & 0xff)

    r.PieceNum += 1
    if r.PieceNum == r.Config.DataPieceCount {
        // We have enough pieces; create a reed-solomon
        // encoding of them
        err = r.Enc.Encode(r.Pieces)       
        if err != nil {
            return
        }

        for i := 0; i < len(r.Pieces); i++ {
            _, err = WriteAllOf(r.ResistText, r.Pieces[i], 0)
            if err != nil {
                return
            }
        }

        r.PieceNum = 0
    }

    return
}

func NewRsWriter(config *RsConfig, outer io.Writer) *WorkerWriter {
    enc, err := reedsolomon.New(config.DataPieceCount, config.ParityPieceCount)
    if err != nil {
        panic(err.Error())
    }

    pieces := make([][]byte, config.DataPieceCount + config.ParityPieceCount)
    for i := 0; i < (config.DataPieceCount + config.ParityPieceCount); i++ {
        pieces[i] = make([]byte, config.DataPieceSize + 4) // 4 checksum bytes
    }

    worker := &RsWriterWorker{
        Config: config,
        ResistText: outer,
        Chunk: make([]byte, config.DataPieceSize),
        Pieces: pieces,
        PieceNum: 0,
        CrcTab: crc32.MakeTable(crc32.Castagnoli),
        Enc: enc,
    }

    return NewWorkerWriter(worker)
}

