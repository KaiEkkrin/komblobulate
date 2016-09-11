package komblobulate

import (
	"fmt"
	"math/rand"
	"testing"
)

// Some test data

func getShortData() []byte {
	return []byte("Nobody expects the Spanish Inquisition!")
}

func getRandomData(data []byte) {
	rng := rand.New(rand.NewSource(int64(9000)))
	n, err := rng.Read(data)
	if err != nil {
		panic(err.Error())
	} else if n != len(data) {
		panic("Wrong number of random bytes read")
	}
}

// Test functions...
// TODO: Test with underlying files that break :)
// (To exercise error propagation)

func TestNullNullShort(t *testing.T) {
	testWriteAndRead(t, getShortData(), &DuringTestNull{}, ResistType_None, CipherType_None, &TestNullNullParams{})
}

func TestNullNullLong(t *testing.T) {
	data := make([]byte, 1024*128)
	getRandomData(data)
	testWriteAndRead(t, data, &DuringTestNull{}, ResistType_None, CipherType_None, &TestNullNullParams{})
}

func TestNullNullCorruptConfig0(t *testing.T) {
	dt := &DuringTestCorruptFile{
		Offsets: []int{1},
		Data:    []byte{6},
	}
	testWriteAndRead(t, getShortData(), dt, ResistType_None, CipherType_None, &TestNullNullParams{})
}

func TestNullNullCorruptConfig1(t *testing.T) {
	dt := &DuringTestCorruptFile{
		Offsets: []int{3*ConfigSize + 1},
		Data:    []byte{6},
	}
	testWriteAndRead(t, getShortData(), dt, ResistType_None, CipherType_None, &TestNullNullParams{})
}

func TestNullNullCorruptConfig2(t *testing.T) {
	dt := &DuringTestCorruptFile{
		Offsets: []int{1},
		Data:    []byte{6},
		FromEnd: true,
	}
	testWriteAndRead(t, getShortData(), dt, ResistType_None, CipherType_None, &TestNullNullParams{})
}

// This one corrupts the body data rather than the config,
// which isn't protected against here:
func TestNullNullCorruptData(t *testing.T) {
	dt := &DuringTestCorruptFile{
		Offsets:           []int{1 + 3*ConfigSize},
		Data:              []byte{6},
		FromEnd:           true,
		CorruptsInnerData: true,
	}
	testWriteAndRead(t, getShortData(), dt, ResistType_None, CipherType_None, &TestNullNullParams{})
}

func TestNullAeadShort(t *testing.T) {
	testWriteAndRead(t, getShortData(), &DuringTestNull{}, ResistType_None, CipherType_Aead, &TestNullAeadParams{256, "password1"})
}

func TestNullAeadLong(t *testing.T) {
	data := make([]byte, 1024*128)
	getRandomData(data)
	testWriteAndRead(t, data, &DuringTestNull{}, ResistType_None, CipherType_Aead, &TestNullAeadParams{256, "password1"})
}

// The AEAD should detect a single corrupt bit,
// but it will panic, because we don't have a good way
// of not doing that :)
// TODO : This still throws the panic out, not sure
// why or how to stop it.  It *was* in a different
// goroutine.  Hmm...
func TODORemoveWhenFixed_TestNullAeadCorruptData(t *testing.T) {
	dt := &DuringTestCorruptFile{
		Offsets: []int{1 + 3*ConfigSize},
		Data:    []byte{6},
		FromEnd: true,
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("MAC failure not detected")
		}
	}()

	testWriteAndRead(t, getShortData(), dt, ResistType_None, CipherType_Aead, &TestNullAeadParams{256, "password1"})
}

// TODO Test short and long chunk sizes, writes several chunks long, etc.

func TestRsNullShort_16_4_2(t *testing.T) {
	testWriteAndRead(t, getShortData(), &DuringTestNull{}, ResistType_Rs, CipherType_None, &TestRsNullParams{16, 4, 2})
}

func TestRsAeadShort_16_4_2(t *testing.T) {
	testWriteAndRead(t, getShortData(), &DuringTestNull{}, ResistType_Rs, CipherType_Aead, &TestRsAeadParams{16, 4, 2, 256, "password1"})
}

func TestRsNullLengths_16_4_2(t *testing.T) {
	params := [][]int{
		[]int{64, 2, 1},
		[]int{16, 4, 2},
		[]int{21, 17, 3},
		// TODO Why is this really slow?  Not that I intended to use it anyway
		//[]int{984, 21, 4},
	}

	for p := 0; p < len(params); p++ {
		fmt.Printf("Using rs params %v\n", params[p])

		for i := 0; i < 6; i++ {
			for sign := -1; sign <= 1; sign += 2 {
				dataLen := 1024*128 + sign*i*i
				fmt.Printf("Testing data length %d...\n", dataLen)

				data := make([]byte, dataLen)
				getRandomData(data)
				testWriteAndRead(t, data, &DuringTestNull{}, ResistType_Rs, CipherType_None, &TestRsNullParams{params[p][0], params[p][1], params[p][2]})
			}
		}
	}
}
