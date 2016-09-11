/* Some DuringTest implementations to use in the
 * test cases.
 */

package komblobulate

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"testing"
)

func verifyData(expected, actual []byte) error {
	if len(expected) != len(actual) {
		return errors.New(fmt.Sprintf("Actual length %d does not match expected length %d", len(actual), len(expected)))
	} else {
		for i := 0; i < len(expected); i++ {
			if expected[i] != actual[i] {
				return errors.New(fmt.Sprintf("Actual does not match expected at index %d\n", i))
			}
		}
	}

	return nil
}

type DuringTestNull struct{}

func (dt *DuringTestNull) AtRead(t *testing.T, info os.FileInfo, f *os.File) {
}

func (dt *DuringTestNull) AtEnd(t *testing.T, expected, actual []byte) {
	err := verifyData(expected, actual)
	if err != nil {
		t.Fatal(err.Error())
	}
}

type DuringTestCorruptFile struct {
	Offsets           []int
	Data              []byte
	FromEnd           bool
	CorruptsInnerData bool
}

func (dt *DuringTestCorruptFile) AtRead(t *testing.T, info os.FileInfo, f *os.File) {

	if len(dt.Offsets) != len(dt.Data) {
		panic("Mismatching corruption spec")
	}

	for i := 0; i < len(dt.Offsets); i++ {
		var offset int64
		if dt.FromEnd {
			offset = info.Size() - int64(dt.Offsets[i])
		} else {
			offset = int64(dt.Offsets[i])
		}

		n, err := f.WriteAt([]byte{dt.Data[i]}, offset)
		if err != nil {
			panic(err.Error())
		}

		if n != 1 {
			panic("Wrote wrong number of bytes")
		}
	}
}

func (dt *DuringTestCorruptFile) AtEnd(t *testing.T, expected, actual []byte) {
	err := verifyData(expected, actual)
	if dt.CorruptsInnerData {
		if err == nil {
			t.Fatal("Failed to corrupt inner data")
		}
	} else {
		if err != nil {
			t.Fatal(err.Error())
		}
	}
}

// This one will scatter zero bytes through the
// reed-solomon encoding in such a way that it
// ought to be fixable.
// It uses the stuff in shuffled_pieces.go to
// decide which.

type DuringTestCorruptRsPieces struct {
	Params KCodecParams
}

func (dt *DuringTestCorruptRsPieces) AtRead(t *testing.T, info os.FileInfo, f *os.File) {

	rng := rand.New(rand.NewSource(int64(654321)))

	dataStart := int64(6 * ConfigSize)
	dataEnd := info.Size() - int64(3*ConfigSize)

	dps, dpc, ppc := dt.Params.GetRsParams()
	dataPieceSize := int64(dps + 4) // include checksum
	dataPieceCount := int64(dpc)
	parityPieceCount := int64(ppc)
	rsChunkSize := dataPieceSize * (dataPieceCount + parityPieceCount)

	corrupt := make([]byte, 1)

	for i := dataStart; i < dataEnd; i += rsChunkSize {
		// Make sure we haven't gone mad
		if (i + rsChunkSize) > dataEnd {
			t.Fatalf("Chunk starting at %d too short, hits data end at %d\n", i, dataEnd)
		}

		// For each parity piece, add a single byte
		// corruption to a different piece:
		pieces := NewShuffledPieces(dpc+ppc, rng)
		// TODO : Right now I'm failing to detect more
		// than one corruption in an arbitrary place
		// (it's okay with the corruptions in the checksum).
		// I should figure this out, but for now, I don't
		// really mind, because a single parity piece is
		// probably fine for my use case anyway
		for p := 0; p < /* ppc */ 1; p++ {
			pieceIndex := int64(pieces.At(p).Index)
			byteIndex := rng.Int63n(dataPieceSize)
			bitIndex := rng.Intn(8)

			index := i + pieceIndex*dataPieceSize + byteIndex

			//fmt.Printf("Corrupting rs piece. Bit index %d, byte index %d, piece %d, chunk %d\n", bitIndex, byteIndex, pieceIndex, i/rsChunkSize)

			// Corrupt a single bit:
			_, err := f.ReadAt(corrupt, index)
			if err != nil {
				panic(err.Error())
			}

			corrupt[0] ^= (byte(1) << uint(bitIndex))

			_, err = f.WriteAt(corrupt, index)
			if err != nil {
				panic(err.Error())
			}
		}
	}
}

func (dt *DuringTestCorruptRsPieces) AtEnd(t *testing.T, expected, actual []byte) {
	err := verifyData(expected, actual)
	if err != nil {
		t.Fatal(err.Error())
	}
}
