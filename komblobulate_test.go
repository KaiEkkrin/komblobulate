package komblobulate

import (    
    "bytes"
    "io"
    "io/ioutil"
    "math/rand"
    "os"
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

// Test utilities...

func verifyData(t *testing.T, expected, actual []byte) {
    if len(expected) != len(actual) {
        t.Fatalf("Actual length %d does not match expected length %d", len(actual), len(expected))
    } else {
        for i := 0; i < len(expected); i++ {
            if expected[i] != actual[i] {
                t.Fatalf("Actual does not match expected at index %d\n", i)
            }
        }
    }
}

func writeTestData(source io.Reader, dest io.Writer, resistType, cipherType byte) (int64, error) {
    writer, err := NewWriter(dest, resistType, cipherType)
    if err != nil {
        return 0, err
    }

    defer writer.Close()

    n, err := io.Copy(writer, source)
    if err != nil {
        return n, err
    }

    return n, nil
}

func testWriteAndRead(t *testing.T, data []byte) {
    input := bytes.NewReader(data)
    
    kblob, err := ioutil.TempFile("", "")
    if err != nil {
        t.Fatal(err.Error())
    }
    blobName := kblob.Name()
    defer func() {
        kblob.Close()
        os.Remove(blobName)
    }()

    n, err := writeTestData(input, kblob, ResistType_None, CipherType_None)
    if err != nil {
        t.Fatal(err.Error())
    }

    if n != int64(len(data)) {
        t.Fatalf("Wrote %d bytes, expected %d", n, len(data))
    }

    // Flush that file out and reopen it:
    kblob.Close()
    kinfo, err := os.Stat(blobName)
    if err != nil {
        t.Fatal(err.Error())
    }

    blobLen := int64(kinfo.Size())
    if blobLen <= int64(len(data)) {
        t.Fatalf("Blob too short, expected more than %d, got %d", len(data), blobLen)
    }

    kblob, err = os.Open(blobName)
    if err != nil {
        t.Fatal(err.Error())
    }

    output := new(bytes.Buffer)
    reader, err := NewReader(kblob, blobLen)
    if err != nil {
        t.Fatal(err.Error())
    }

    n, err = io.Copy(output, reader)
    if err != nil {
        t.Fatal(err.Error())
    }

    if n != int64(len(data)) {
        t.Fatalf("Read %d bytes, expected %d", n, len(data))
    }

    verifyData(t, data, output.Bytes())
}

// Test functions...

func TestNullNullShort(t *testing.T) {
    testWriteAndRead(t, getShortData())
}

func TestNullNullLong(t *testing.T) {
    data := make([]byte, 1024 * 128)
    getRandomData(data)
    testWriteAndRead(t, data)
}

