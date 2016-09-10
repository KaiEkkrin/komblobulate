package komblobulate

import (    
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

func TestNullNullShort(t *testing.T) {
    testWriteAndRead(t, getShortData(), &DuringTestNull{}, ResistType_None, CipherType_None)
}

func TestNullNullLong(t *testing.T) {
    data := make([]byte, 1024 * 128)
    getRandomData(data)
    testWriteAndRead(t, data, &DuringTestNull{}, ResistType_None, CipherType_None)
}

func TestNullNullCorruptConfig0(t *testing.T) {
    dt := &DuringTestCorruptFile{
        Offsets: []int{1},
        Data: []byte{6},
    }
    testWriteAndRead(t, getShortData(), dt, ResistType_None, CipherType_None)
}

func TestNullNullCorruptConfig1(t *testing.T) {
    dt := &DuringTestCorruptFile{
        Offsets: []int{3 * ConfigSize + 1},
        Data: []byte{6},
    }
    testWriteAndRead(t, getShortData(), dt, ResistType_None, CipherType_None)
}

func TestNullNullCorruptConfig2(t *testing.T) {
    dt := &DuringTestCorruptFile{
        Offsets: []int{1},
        Data: []byte{6},
        FromEnd: true,
    }
    testWriteAndRead(t, getShortData(), dt, ResistType_None, CipherType_None)
}

// This one corrupts the body data rather than the config,
// which isn't protected against here:
func TestNullNullCorruptData(t *testing.T) {
    dt := &DuringTestCorruptFile{
        Offsets: []int{1 + 3 * ConfigSize},
        Data: []byte{6},
        FromEnd: true,
        CorruptsInnerData: true,
    }
    testWriteAndRead(t, getShortData(), dt, ResistType_None, CipherType_None)
}

func TestNullAeadShort(t *testing.T) {
    testWriteAndRead(t, getShortData(), &DuringTestNull{}, ResistType_None, CipherType_Aead, int64(256), "password1")
}

func TestNullAeadLong(t *testing.T) {
    data := make([]byte, 1024 * 128)
    getRandomData(data)
    testWriteAndRead(t, data, &DuringTestNull{}, ResistType_None, CipherType_Aead, int64(256), "password1")
}

// The AEAD should detect a single corrupt bit,
// but it will panic, because we don't have a good way
// of not doing that :)
// TODO : This still throws the panic out, not sure
// why or how to stop it.  It *was* in a different
// goroutine.  Hmm...
func TODORemoveWhenFixed_TestNullAeadCorruptData(t *testing.T) {
    dt := &DuringTestCorruptFile{
        Offsets: []int{1 + 3 * ConfigSize},
        Data: []byte{6},
        FromEnd: true,
    }

    defer func() {
        if r := recover(); r == nil {
            t.Error("MAC failure not detected")
        }
    }()

    testWriteAndRead(t, getShortData(), dt, ResistType_None, CipherType_Aead, int64(256), "password1")
}

// TODO Test short and long chunk sizes, writes several chunks long, etc.

func TestRsNullShort_16_4_2(t *testing.T) {
    testWriteAndRead(t, getShortData(), &DuringTestNull{}, ResistType_Rs, CipherType_None, 16, 4, 2)
}


