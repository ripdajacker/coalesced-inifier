package read

import (
	"bytes"
	"coalesced-inifier/gowenc"
	"encoding/binary"
	"fmt"
	"strings"
)

type BinaryIniKeyValue struct {
	Key   string
	Value string
}

type BinaryIniSection struct {
	Name   string
	Values []BinaryIniKeyValue
}

type BinaryIniFile struct {
	Name     string
	Sections []BinaryIniSection
}

type BinaryCoalescedIniFiles struct {
	fileCount uint32
	Files     []BinaryIniFile
}

func readIniSection(r *bytes.Reader) (*BinaryIniSection, error) {
	var result BinaryIniSection

	sectionName, err := gowenc.ReadUtf16InvLen(r)
	if err != nil {
		return nil, err
	}

	result.Name = sectionName

	valueCount, err := gowenc.ReadUint32BE(r)
	if err != nil {
		return nil, err
	}

	result.Values = make([]BinaryIniKeyValue, valueCount)
	for i := 0; i < int(valueCount); i++ {
		value, err := readIniValue(r)
		if err != nil {
			return nil, err
		}

		result.Values[i] = *value
	}

	return &result, nil
}

func readIniValue(r *bytes.Reader) (*BinaryIniKeyValue, error) {
	var result BinaryIniKeyValue

	key, err := gowenc.ReadUtf16InvLen(r)
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(key, ";") {
		key = strings.Replace(key, ";", "IGNORED_SC_", 1)
	} else if strings.HasPrefix(key, "#") {
		key = strings.Replace(key, "#", "IGNORED_SH_", 1)
	}

	value, err := gowenc.ReadUtf16InvLen(r)
	if err != nil {
		return nil, err
	}

	result.Key = key

	if value == "\\\\\\\\" {
		value = fmt.Sprintf("`%s`", value)
	}
	result.Value = value

	return &result, nil
}

func readIniFile(r *bytes.Reader) (*BinaryIniFile, error) {
	var result BinaryIniFile

	fileName, err := gowenc.ReadUtf16InvLen(r)
	if err != nil {
		return nil, err
	}

	result.Name = fileName

	sectionCount, err := gowenc.ReadUint32BE(r)
	fmt.Printf("Reading %s with %d sections\n", fileName, sectionCount)

	if err != nil {
		return nil, err
	}

	result.Sections = make([]BinaryIniSection, sectionCount, sectionCount)
	for i := uint32(0); i < sectionCount; i++ {
		section, err := readIniSection(r)
		if err != nil {
			return nil, err
		}

		result.Sections[i] = *section

	}

	return &result, nil
}

func ReadCoalescedIniFiles(r *bytes.Reader) (*BinaryCoalescedIniFiles, error) {
	var result BinaryCoalescedIniFiles

	err := binary.Read(r, binary.BigEndian, &result.fileCount)
	if err != nil {
		panic(err)
	}

	result.Files = make([]BinaryIniFile, result.fileCount, result.fileCount)

	for i := uint32(0); i < result.fileCount; i++ {
		iniFile, err := readIniFile(r)
		if err != nil {
			return nil, err
		}

		result.Files[i] = *iniFile
	}

	return &result, nil
}
