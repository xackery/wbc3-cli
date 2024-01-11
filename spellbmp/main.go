package main

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/bmp"
)

func main() {
	err := run()
	if err != nil {
		fmt.Println("Failed to run:", err)
		os.Exit(1)
	}
}

func run() error {

	if len(os.Args) < 3 {
		fmt.Println("usage: spellbmp <inputdir> <outputdir>")
		os.Exit(1)
	}

	inputDir := os.Args[1]
	outputDir := os.Args[2]

	dir, err := os.Open(inputDir)
	if err != nil {
		return fmt.Errorf("open input dir: %w", err)
	}

	files, err := dir.Readdir(0)
	if err != nil {
		return fmt.Errorf("read input dir: %w", err)
	}

	err = os.RemoveAll(outputDir)
	if err != nil {
		return fmt.Errorf("remove output dir: %w", err)
	}

	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	for _, file := range files {

		ext := filepath.Ext(file.Name())
		// if it's a bmp
		if file.IsDir() || ext != ".bmp" {
			continue
		}

		if strings.HasSuffix(file.Name(), "n.bmp") {
			continue
		}

		outName := strings.TrimSuffix(file.Name(), ext)

		err = convert(filepath.Join(inputDir, file.Name()), filepath.Join(outputDir, outName+".png"))
		if err != nil {
			return fmt.Errorf("convert %s: %w", file.Name(), err)
		}
	}

	return nil
}

// convert takes a bmp file and turns it into a png
func convert(in string, out string) error {
	r, err := os.Open(in)
	if err != nil {
		return fmt.Errorf("convert open: %w", err)
	}
	defer r.Close()

	w, err := os.Create(out)
	if err != nil {
		return fmt.Errorf("convert create: %w", err)
	}
	defer w.Close()

	dec, err := bmp.Decode(r)
	if err != nil {
		return fmt.Errorf("convert decode: %w", err)
	}

	err = png.Encode(w, dec)
	if err != nil {
		return fmt.Errorf("convert encode: %w", err)
	}

	return nil
}
