package gow2

import (
	"bytes"
	"coalesced-inifier/gow2/read"
	"coalesced-inifier/gow2/write"
	"coalesced-inifier/gowenc"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/ini.v1"
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
)

func writeInitFileToDisk(file read.BinaryIniFile, baseDir string, prefix string) (string, error) {
	realFileName := strings.Replace(file.Name, prefix, "", 1)
	realFileName = strings.Replace(realFileName, "\\", "/", -1)

	dir, _ := path.Split(path.Join(baseDir, realFileName))
	err := os.MkdirAll(dir, 0755)

	if err != nil {
		panic(err)
	}

	join := path.Join(baseDir, realFileName)
	outputFile, err := os.Create(join)
	if err != nil {
		panic(err)
	}

	defer outputFile.Close()

	iniOut := ini.Empty(ini.LoadOptions{AllowShadows: true})
	for _, binarySection := range file.Sections {
		section := iniOut.Section(binarySection.Name)

		for _, values := range binarySection.Values {
			_, err := section.NewKey(values.Key, values.Value)
			if err != nil {
				return "", err
			}
		}
	}

	_, err = iniOut.WriteTo(outputFile)
	return realFileName, err
}

func detectPrefix(r *bytes.Reader) (*string, error) {
	coalesced, err := read.ReadCoalescedIniFiles(r)
	if err != nil {
		return nil, err
	}

	prefixMap := make(map[string]bool)
	for i := 0; i < len(coalesced.Files); i++ {
		file := coalesced.Files[i]

		splitAt := -1
		for i, char := range file.Name {
			if char != '\\' && char != '.' {
				splitAt = i
				break
			}
		}

		prefixMap[file.Name[:splitAt]] = true
	}

	keys := reflect.ValueOf(prefixMap).MapKeys()
	if len(keys) != 1 {
		return nil, errors.New("more than one prefix found")
	}

	result := keys[0].String()
	return &result, nil

}

func Unpack(r *bytes.Reader, outputDir string) error {
	coalescedBin, err := io.ReadAll(r)
	prefix, err := detectPrefix(bytes.NewReader(coalescedBin))
	if err != nil {
		return err
	}

	metadata, err := json.Marshal(Metadata{Prefix: *prefix})
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(outputDir, "metadata.json"), metadata, 0644)
	if err != nil {
		return err
	}

	coalesced, err := read.ReadCoalescedIniFiles(bytes.NewReader(coalescedBin))
	if err != nil {
		return err
	}

	for i := 0; i < len(coalesced.Files); i++ {
		file := coalesced.Files[i]

		fmt.Printf("Writing '%s' to disk...", file.Name)
		realFileName, err := writeInitFileToDisk(file, outputDir, *prefix)

		if err != nil {
			return err
		}

		fmt.Printf(" written to '%s'.\n", realFileName)
	}

	fmt.Printf("Unpacked %d files to '%s'.\n", len(coalesced.Files), outputDir)
	return nil
}

func Pack(inputFiles []string, baseDir string, prefix string) (*bytes.Buffer, error) {
	w := bytes.Buffer{}

	err := gowenc.WriteUint32BE(&w, len(inputFiles))

	if err != nil {
		return nil, err
	}

	for _, inputFile := range inputFiles {
		err = write.AppendIni(&w, baseDir, inputFile, prefix)
		if err != nil {
			return nil, err
		}

		fmt.Printf(" done.\n")
	}

	hash := md5.Sum(w.Bytes())

	println()
	fmt.Printf("File length: %d\n", len(w.Bytes()))
	fmt.Printf("  File hash: %s\n\n", hex.EncodeToString(hash[:]))

	return &w, nil
}

type Metadata struct {
	Prefix string `json:"prefix"`
}
