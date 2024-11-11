package gowenc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"unicode/utf16"
)

func ReadUtf16InvLen(r *bytes.Reader) (string, error) {
	var length uint32

	err := binary.Read(r, binary.BigEndian, &length)
	if err != nil {
		return "", err
	}

	// There's a special case for missing values == 0x0 even if inverted
	if length == 0 || length == 0xFFFFFFFF {
		return "", nil
	}

	length = ^length

	// We are reading UTF-16 strings with a 2-byte nul string terminator
	b := make([]byte, length*2+2)
	n, err := r.Read(b)

	if err != nil {
		return "", err
	}

	if uint32(n) != length*2+2 {
		panic(fmt.Sprintf("Read %d bytes, expected %d", n, length*2+2))
	}

	chars := make([]uint16, length)
	for i := 0; i < int(length); i++ {
		chars[i] = uint16(b[i*2]) | (uint16(b[(i*2)+1]) << 8)
	}

	runes := utf16.Decode(chars)
	return string(runes), nil
}

func ReadUint32BE(r *bytes.Reader) (uint32, error) {
	var length uint32
	err := binary.Read(r, binary.BigEndian, &length)
	return length, err
}

func WriteUint32BE(w *bytes.Buffer, v int) error {
	return binary.Write(w, binary.BigEndian, uint32(v))
}

func WriteUtf16InvLen(w *bytes.Buffer, s string) error {
	bb := []byte(s)
	rns := bytes.Runes(bb)

	err1 := WriteUint32BE(w, ^len(rns))
	for _, b := range rns {
		err2 := w.WriteByte(byte(b & 0xFF))
		err3 := w.WriteByte(byte((b >> 8) & 0xFF))

		if err := errors.Join(err2, err3); err != nil {
			return err
		}
	}

	err4 := w.WriteByte(0)
	err5 := w.WriteByte(0)
	return errors.Join(err1, err4, err5)
}
