package komblobulate

import (
	"github.com/klauspost/reedsolomon"
	"hash/crc32"
	"io"
)

type RsReaderWorker struct {
	Config *RsConfig

	ResistText io.Reader

	Pieces        [][]byte
	PieceNum      int
	MissingPieces int

	// This tracks how much data we've decoded
	InnerLengthRead int64

	CrcTab *crc32.Table
	Enc    reedsolomon.Encoder
}

func (r *RsReaderWorker) Ready(putChunk func([]byte) error) (err error) {

	dataPieceSize := int(r.Config.DataPieceSize)
	dataPieceCount := int(r.Config.DataPieceCount)
	parityPieceCount := int(r.Config.ParityPieceCount)

	// Read the next chunk.
	_, err = ReadAllOf(r.ResistText, r.Pieces[r.PieceNum], 0)
	if err != nil {
		return
	}

	// If this is a data piece, verify its checksum.
	if r.PieceNum < dataPieceCount {
		writtenChecksum := uint32(r.Pieces[r.PieceNum][dataPieceSize])
		writtenChecksum |= (uint32(r.Pieces[r.PieceNum][dataPieceSize+1]) << 8)
		writtenChecksum |= (uint32(r.Pieces[r.PieceNum][dataPieceSize+2]) << 16)
		writtenChecksum |= (uint32(r.Pieces[r.PieceNum][dataPieceSize+3]) << 24)

		calcChecksum := crc32.Checksum(r.Pieces[r.PieceNum][:dataPieceSize], r.CrcTab)
		if calcChecksum != writtenChecksum {
			// This piece isn't valid.  Drop it; we'll use
			// the reed-solomon encoding to reconstruct it,
			// hopefully.
			r.MissingPieces += 1
			r.Pieces[r.PieceNum] = nil
		}
	}

	r.PieceNum += 1
	if r.PieceNum == (dataPieceCount + parityPieceCount) {

		// Reconstruct any missing pieces:
		if r.MissingPieces > 0 {
			err = r.Enc.Reconstruct(r.Pieces)
			if err != nil {
				return
			}
		}

		// Push the data pieces into the next tier
		// of reading:
		for i := 0; i < dataPieceCount; i++ {
			// The inner data will have been padded to fit
			// the final reed-solomon matrix; track this
			// and drop the final padding
			pieceLength := int64(dataPieceSize)
			if (r.InnerLengthRead + pieceLength) > r.Config.TotalInnerLength {
				pieceLength = r.Config.TotalInnerLength - r.InnerLengthRead
			}

			if pieceLength > 0 {
				err = putChunk(r.Pieces[i][:pieceLength])
				if err != nil {
					return
				}
			}

			r.InnerLengthRead += pieceLength
		}

		// Reset things for the next set of pieces:
		r.MissingPieces = 0
		r.PieceNum = 0
	}

	return
}

func NewRsReader(config *RsConfig, outer io.Reader) *WorkerReader {
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

	worker := &RsReaderWorker{
		Config:          config,
		ResistText:      outer,
		Pieces:          pieces,
		PieceNum:        0,
		MissingPieces:   0,
		InnerLengthRead: 0,
		CrcTab:          crc32.MakeTable(crc32.Castagnoli),
		Enc:             enc,
	}

	return NewWorkerReader(worker)
}
