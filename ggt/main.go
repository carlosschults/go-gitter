package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	arguments := os.Args
	path, _ := filepath.Abs(".git")
	standardizedPath := filepath.ToSlash(path) + "/"
	fmt.Println()

	// hash-object
	if len(arguments) >= 1 && arguments[1] == "hash-object" {
		runHashObjectCommand(arguments)
	}

	// cat-file
	if len(arguments) >= 1 && arguments[1] == "cat-file" {
		runCatFileCommand(arguments)
	}

	// init
	if len(arguments) == 2 && arguments[1] == "init" {

		if _, err := os.Stat(".git"); err == nil {
			fmt.Println("Reinitialized existing Git repository in", standardizedPath)
			os.Exit(0)
		}

		if err := os.Mkdir(".git", os.ModePerm); err != nil {
			log.Fatal(err)
		}

		if err := os.WriteFile(".git/HEAD", []byte("ref: refs/heads/main"), 0666); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}

		if err := os.MkdirAll(".git/objects/info", os.ModePerm); err != nil {
			log.Fatal(err)
		}

		if err := os.MkdirAll(".git/objects/pack", os.ModePerm); err != nil {
			log.Fatal(err)
		}

		if err := os.MkdirAll(".git/refs/heads", os.ModePerm); err != nil {
			log.Fatal(err)
		}

		if err := os.MkdirAll(".git/refs/tags", os.ModePerm); err != nil {
			log.Fatal(err)
		}

		fmt.Println("Initialized empty Git repository in", standardizedPath)
		os.Exit(0)
	}

	fmt.Println("Unknown command")
	os.Exit(1)
}

func runCatFileCommand(arguments []string) {
	if len(arguments) != 4 || arguments[2] != "-p" {
		return
	}
	h := arguments[3]
	folderName := h[0:2]
	fileName := h[2:]
	fullPath := filepath.Join(".git/objects", folderName, fileName)

	var contents []byte
	var err error
	contents, err = os.ReadFile(fullPath)

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	r, err := zlib.NewReader(bytes.NewReader(contents))
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	buf := new(strings.Builder)
	_, err = io.Copy(buf, r)

	if err != nil {
		log.Fatal(err)
	}

	uncompressedContents := buf.String()
	contentsWithoutHeader := strings.Split(uncompressedContents, "\x00")[1]

	r.Close()
	fmt.Println(contentsWithoutHeader)
	os.Exit(0)
}

func runHashObjectCommand(arguments []string) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Printf("failed to read data: %v", err)
		return
	}

	saveFile := len(arguments) == 4 && arguments[2] == "-w"

	size := len(data)
	header := fmt.Sprintf("blob %s%c", strconv.Itoa(size), 0)
	content := header + string(data)
	hash := sha1.New()
	hash.Write([]byte(content))
	hashedData := hash.Sum(nil)
	hashedString := hex.EncodeToString(hashedData)
	folderName := hashedString[0:2]
	fileName := hashedString[2:]

	if saveFile {
		// create the directory for the blob
		if err := os.Mkdir(".git/objects/"+folderName, os.ModePerm); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}

		// compress the content using zlib and save the file
		var buffer bytes.Buffer
		w := zlib.NewWriter(&buffer)
		w.Write([]byte(content))
		w.Close()
		if err := os.WriteFile(".git/objects/"+folderName+"/"+fileName, buffer.Bytes(), 0666); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	}

	fmt.Println(hashedString)
	os.Exit(0)
}
