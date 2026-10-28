package marker

import "testing"

func TestMarkerEncodings(t *testing.T) {
	checkRuneAndByteSlice(MarkerEscaping, BytesMarkerEscaping)
	checkRuneAndByteSlice(MarkerCodeBlockNewline, BytesMarkerCodeBlockNewline)
	defer func() {
		if recover() == nil {
			t.Fatal("mismatched marker encoding did not panic")
		}
	}()
	checkRuneAndByteSlice('x', []byte("y"))
}
