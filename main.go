package main

import (
	"bytes"
	"coalesced-inifier/gow2"
	"coalesced-inifier/web"
	"fmt"
	"github.com/alexflint/go-arg"
	"os"
	"path"
	"strings"
)

type PackCommand struct {
	Input  string `arg:"-i,--input,required" placeholder:"<ini folder>" help:"The path containing your INI files"`
	Output string `arg:"-o,--output,required" placeholder:"<coalesced bin>" help:"The coalesced bin file such as Coalesced_int.bin"`
	Game   string `arg:"-g,--game,required" help:"game to pack data for, one of: gow2, lollipop"`
}

type UnpackCommand struct {
	Input  string `arg:"-i,--input,required" placeholder:"<coalesced bin>" help:"The coalesced bin file such as Coalesced_int.bin"`
	Output string `arg:"-o,--output,required" placeholder:"<ini folder>" help:"The path containing your INI files"`
	Game   string `arg:"-g,--game,required" help:"game to unpack data from, one of: gow2, lollipop"`
}

type WebCommand struct {
	Port int16 `arg:"-p,--port" default:"8080" help:"The port to listen on"`
}

type args struct {
	WebCommand    *WebCommand    `arg:"subcommand:web" help:"Start web server"`
	PackCommand   *PackCommand   `arg:"subcommand:pack" help:"Pack a folder containing INI files to a coalesced bin file"`
	UnpackCommand *UnpackCommand `arg:"subcommand:unpack" help:"Unpack a coalesced bin file to a folder containing INI files to"`
}

func (args) Version() string {
	return "0.1.0"
}

func (args) Description() string {
	return "Coalesced Inifier"
}

func main() {
	var args args
	arg.MustParse(&args)

	var err error
	if args.PackCommand != nil {
		err = pack(args.PackCommand)
	} else if args.UnpackCommand != nil {
		err = unpack(args.UnpackCommand)
	} else if args.WebCommand != nil {
		err = web.Hook(args.WebCommand.Port)
	}

	if err != nil {
		panic(err)
	}
}

func pack(cmd *PackCommand) error {
	prefix := parseGamePrefix(cmd.Game)

	fileList := make([]string, 0)
	err := recursiveFileList(cmd.Input, &fileList)

	if err != nil {
		return err
	}

	fmt.Printf("Packing folder '%s' to '%s' using prefix '%s'\n\n", cmd.Input, cmd.Output, prefix)
	buf, err := gow2.Pack(fileList, cmd.Input, prefix)

	if err != nil {
		return err
	}

	outFile, err := os.Create(cmd.Output)
	defer outFile.Close()

	if err != nil {
		return err
	}

	_, err = buf.WriteTo(outFile)

	fmt.Printf("Successfully packed %d files to '%s'.\n", len(fileList), outFile.Name())
	return err
}

func unpack(cmd *UnpackCommand) error {
	prefix := parseGamePrefix(cmd.Game)
	data, err := os.ReadFile(cmd.Input)

	if err != nil {
		panic(err)
	}

	fmt.Printf("Unpacking '%s' to '%s' using prefix '%s'\n\n", cmd.Input, cmd.Output, prefix)
	reader := bytes.NewReader(data)
	return gow2.Unpack(reader, cmd.Output, prefix)
}

func parseGamePrefix(game string) string {
	switch game {
	case "gow2":
		return "..\\"
	case "lollipop":
		return "..\\..\\"
	}
	panic("Invalid game: " + game)
}

func recursiveFileList(current string, inputFiles *[]string) error {
	dir, err := os.ReadDir(current)
	if err != nil {
		return err
	}

	for i := 0; i < len(dir); i++ {
		entry := dir[i]
		entryPath := path.Join(current, entry.Name())
		if entry.IsDir() {
			err = recursiveFileList(entryPath, inputFiles)
		} else if isIniOrInt(entryPath) {
			*inputFiles = append(*inputFiles, entryPath)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func isIniOrInt(file string) bool {
	file = strings.ToLower(file)
	return strings.HasSuffix(file, ".int") || strings.HasSuffix(file, ".ini")
}
