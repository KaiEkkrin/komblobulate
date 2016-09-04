/* Some DuringTest implementations to use in the
 * test cases.
*/

package komblobulate

import (
    "errors"
    "fmt"
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

type DuringTestNull struct {}

func (dt *DuringTestNull) AtRead(t *testing.T, info os.FileInfo, f *os.File) {
}

func (dt *DuringTestNull) AtEnd(t *testing.T, expected, actual []byte) {
    err := verifyData(expected, actual)
    if err != nil {
        t.Fatal(err.Error())
    }
}

type DuringTestCorruptFile struct {
    Offsets []int
    Data []byte
    FromEnd bool
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

