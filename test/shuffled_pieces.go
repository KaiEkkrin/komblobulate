/* This helper for duringtest.go shuffles a list of
 * the pieces in a reed-solomon encoded chunk.
 */

package test

import (
	"math/rand"
	"sort"
)

type ShuffledPiece struct {
	Index int
	Order int
}

type ShuffledPieces struct {
	Pieces []ShuffledPiece
}

func (p *ShuffledPieces) At(i int) ShuffledPiece {
	return p.Pieces[i]
}

func (p *ShuffledPieces) Len() int {
	return len(p.Pieces)
}

func (p *ShuffledPieces) Less(i, j int) bool {
	return p.Pieces[i].Order < p.Pieces[j].Order
}

func (p *ShuffledPieces) Swap(i, j int) {
	sw := p.Pieces[i]
	p.Pieces[i] = p.Pieces[j]
	p.Pieces[j] = sw
}

func NewShuffledPieces(count int, rng *rand.Rand) *ShuffledPieces {
	ps := make([]ShuffledPiece, count)
	for i := 0; i < count; i++ {
		ps[i].Index = i
		ps[i].Order = rng.Int()
	}

	pieces := &ShuffledPieces{ps}
	sort.Sort(pieces)
	return pieces
}
