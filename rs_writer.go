package komblobulate

// TODO Try using github.com/klauspost/crc32 for
// a large speedup, apparently
import (
	"github.com/klauspost/reedsolomon"
	"hash/crc32"
	"io"
)

type RsWriterWorker struct {
	Config *RsConfig

	ResistText io.Writer

	Chunk    []byte
	Pieces   [][]byte
	PieceNum int

	CrcTab *crc32.Table
	Enc    reedsolomon.Encoder
}

func (r *RsWriterWorker) includeChunk() (err error) {
	dataPieceSize := int(r.Config.DataPieceSize)
	dataPieceCount := int(r.Config.DataPieceCount)

	// Checksum the current chunk, and write it and the checksum
	// into the next piece:
	checksum := crc32.Checksum(r.Chunk, r.CrcTab)

	copy(r.Pieces[r.PieceNum], r.Chunk)
	r.Pieces[r.PieceNum][dataPieceSize] = byte(checksum & 0xff)
	r.Pieces[r.PieceNum][dataPieceSize+1] = byte((checksum >> 8) & 0xff)
	r.Pieces[r.PieceNum][dataPieceSize+2] = byte((checksum >> 16) & 0xff)
	r.Pieces[r.PieceNum][dataPieceSize+3] = byte((checksum >> 24) & 0xff)

	r.PieceNum += 1
	if r.PieceNum == dataPieceCount {
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

func (r *RsWriterWorker) ChunkSize() int {
	return int(r.Config.DataPieceSize)
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

	// Include this chunk:
	return r.includeChunk()
}

func (r *RsWriterWorker) Close() (err error) {
	// While we have trailing pieces, pad with arbitrary
	// chunks.  The pieces will be flushed out once we
	// reach the data piece count, and they'll be ignored
	// on read, because we're not incrementing
	// TotalInnerLength:
	for r.PieceNum > 0 {
		err = r.includeChunk()
		if err != nil {
			return
		}
	}

	return
}

func NewRsWriter(config *RsConfig, outer io.Writer) *WorkerWriter {
	dataPieceSize := int(config.DataPieceSize)
	dataPieceCount := int(config.DataPieceCount)
	parityPieceCount := int(config.ParityPieceCount)

	enc, err := reedsolomon.New(dataPieceCount, parityPieceCount)
	if err != nil {
		panic(err.Error())
	}

	pieces := make([][]byte, dataPieceCount+parityPieceCount)
	for i := 0; i < len(pieces); i++ {
		pieces[i] = make([]byte, dataPieceSize+4) // 4 checksum bytes
	}

	worker := &RsWriterWorker{
		Config:     config,
		ResistText: outer,
		Chunk:      make([]byte, dataPieceSize),
		Pieces:     pieces,
		PieceNum:   0,
		CrcTab:     crc32.MakeTable(crc32.Castagnoli),
		Enc:        enc,
	}

	return NewWorkerWriter(worker)
}
