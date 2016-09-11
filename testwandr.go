/* Core testing function, defining an interface that
 * describes what things to do during the test.
*/

package komblobulate

import (
    "bytes"
    "fmt"
    "io"
    "os"
    "testing"
    )

type DuringTest interface {
    // Called just after opening the written file
    // for read.
    AtRead(*testing.T, os.FileInfo, *os.File)

    // Called after the contained data has been
    // read back.  (expected, actual)
    AtEnd(*testing.T, []byte, []byte)
}

func writeTestData(source io.Reader, dest io.WriteSeeker, resistType, cipherType byte, p ...interface{}) (int64, error) {
    writer, err := NewWriter(dest, resistType, cipherType, p...)
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

func testWriteAndRead(t *testing.T, data []byte, dt DuringTest, resistType, cipherType byte, p ...interface{}) {
    input := bytes.NewReader(data)
    
    kblob, err := os.Create("temp.kblob")
    if err != nil {
        t.Fatal(err.Error())
    }
    blobName := kblob.Name()
    defer func() {
        kblob.Close()
        os.Remove(blobName)
    }()

    n, err := writeTestData(input, kblob, resistType, cipherType, p...)
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

    // TODO remove debug
    fmt.Printf("Kblob length is %d\n", blobLen)

    kblob, err = os.OpenFile(blobName, os.O_RDWR, 0)
    if err != nil {
        t.Fatal(err.Error())
    }

    dt.AtRead(t, kinfo, kblob)

    output := new(bytes.Buffer)
    reader, err := NewReader(kblob, blobLen, p...)
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

    dt.AtEnd(t, data, output.Bytes())
}

