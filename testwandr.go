/* Core testing function, defining an interface that
 * describes what things to do during the test.
*/

package komblobulate

import (
    "bytes"
    "io"
    "io/ioutil"
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

func testWriteAndRead(t *testing.T, data []byte, dt DuringTest) {
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

    kblob, err = os.OpenFile(blobName, os.O_RDWR, 0)
    if err != nil {
        t.Fatal(err.Error())
    }

    dt.AtRead(t, kinfo, kblob)

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

    dt.AtEnd(t, data, output.Bytes())
}

