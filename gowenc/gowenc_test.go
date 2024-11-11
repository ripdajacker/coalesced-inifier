package gowenc

import (
	"bytes"
	"encoding/hex"
	"os"
	"testing"
)

func TestReadWriteStrings(t *testing.T) {
	testSerializationOfString(t, "../testdata/utf16_ascii_string.bin")
	testSerializationOfString(t, "../testdata/utf16_copyright_message.bin")
}

func testSerializationOfString(t *testing.T, file string) {
	expectedContent, err := os.ReadFile(file)
	if err != nil {
		panic(err)
	}

	buffer := bytes.NewReader(expectedContent)
	deserializedString, err := ReadUtf16InvLen(buffer)
	if err != nil {
		panic(err)
	}

	serBuf := bytes.NewBuffer(make([]byte, 0))
	err = WriteUtf16InvLen(serBuf, deserializedString)
	if err != nil {
		panic(err)
	}

	if !bytes.Equal(serBuf.Bytes(), expectedContent) {
		t.Errorf("Expected: %s", hex.EncodeToString(expectedContent))
		t.Errorf("     Got: %s", hex.EncodeToString(serBuf.Bytes()))

		t.Fatal("Serialization mishap")
	}
}
