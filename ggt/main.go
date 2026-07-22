package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/binary"
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

	// update-index
	if len(arguments) >= 1 && arguments[1] == "update-index" {
		runUpdateIndexCommand(arguments)
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
	if len(arguments) != 4 {
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
	parts := strings.Split(uncompressedContents, "\x00")
	flag := arguments[2]
	var result string

	switch flag {
	case "-p":
		result = parts[1]
	case "-t":
		result = strings.Fields(parts[0])[0]
	case "-s":
		result = strings.Fields(parts[0])[1]
	}

	r.Close()
	fmt.Println(result)
	os.Exit(0)
}

func hashData(data []byte, objectType string) (string, []byte) {
	size := len(data)
	header := fmt.Sprintf("%s %s%c", objectType, strconv.Itoa(size), 0)
	contentToBeHashed := header + string(data)
	hash := sha1.New()
	hash.Write([]byte(contentToBeHashed))
	return contentToBeHashed, hash.Sum(nil)
}

func runHashObjectCommand(arguments []string) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Printf("failed to read data: %v", err)
		return
	}

	saveFile := len(arguments) == 4 && arguments[2] == "-w"

	unhashedContent, hashedData := hashData(data, "blob")
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
		w.Write([]byte(unhashedContent))
		w.Close()
		if err := os.WriteFile(".git/objects/"+folderName+"/"+fileName, buffer.Bytes(), 0666); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	}

	fmt.Println(hashedString)
	os.Exit(0)
}

func runUpdateIndexCommand(arguments []string) {

	if len(arguments) != 4 {
		return
	}

	if arguments[2] != "--add" {
		return
	}

	// TODO improve variable names
	dataToSave := []byte{}
	var err error

	// appending DIRC
	dataToSave, err = binary.Append(dataToSave, binary.BigEndian, []byte("DIRC"))
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// appending version number (2)
	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, 2)

	dataToSave, err = binary.Append(dataToSave, binary.BigEndian, bs)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// appending the number of entries (1)
	dataToSave, err = append4ByteBigEndianInteger(dataToSave, 1)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// let's append several pieces of information that will be mostly zeroed-out
	dataToSave, err = append4ByteBigEndianInteger(dataToSave, 0)     // ctime seconds
	dataToSave, err = append4ByteBigEndianInteger(dataToSave, 0)     // ctime nanoseconds
	dataToSave, err = append4ByteBigEndianInteger(dataToSave, 0)     // mtime seconds
	dataToSave, err = append4ByteBigEndianInteger(dataToSave, 0)     // mtime nanoseconds
	dataToSave, err = append4ByteBigEndianInteger(dataToSave, 0)     // dev
	dataToSave, err = append4ByteBigEndianInteger(dataToSave, 0)     // ino
	dataToSave, err = append4ByteBigEndianInteger(dataToSave, 33188) // mode
	dataToSave, err = append4ByteBigEndianInteger(dataToSave, 0)     // uid
	dataToSave, err = append4ByteBigEndianInteger(dataToSave, 0)     // gid

	// append file size
	filePath := arguments[3]
	fi, err := os.Lstat(filePath)
	if err != nil {
		log.Fatal(err)
	}
	size := fi.Size()
	dataToSave, err = append4ByteBigEndianInteger(dataToSave, size) // size

	contents, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// append the hash
	_, hash := hashData(contents, "blob")
	dataToSave, err = binary.Append(dataToSave, binary.BigEndian, hash)

	// flags
	bs2 := make([]byte, 2)
	binary.BigEndian.PutUint16(bs2, uint16(len(filePath)))
	dataToSave, err = binary.Append(dataToSave, binary.BigEndian, bs2)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// filename bytes
	dataToSave, err = binary.Append(dataToSave, binary.BigEndian, []byte(filePath))
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// null byte
	dataToSave, err = binary.Append(dataToSave, binary.BigEndian, []byte{0})
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// adding the padding bytes
	dataLength := len(dataToSave) - 12
	remainder := dataLength % 8

	if remainder > 0 {
		paddingLength := 8 - remainder
		bs2 := make([]byte, paddingLength)
		dataToSave, err = binary.Append(dataToSave, binary.BigEndian, bs2)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	}

	// hash the content and append to the file
	contentHash := sha1.New()
	contentHash.Write(dataToSave)
	dataToSave, err = binary.Append(dataToSave, binary.BigEndian, contentHash.Sum(nil))
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// saving the file
	if err = os.WriteFile(".git/index", dataToSave, 0666); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	os.Exit(0)
}

func append4ByteBigEndianInteger(dataToSave []byte, i int64) ([]byte, error) {
	return appendNByteBigEndianInteger(dataToSave, i, 4)
}

func appendNByteBigEndianInteger(dataToSave []byte, i int64, capacity int) ([]byte, error) {
	bs2 := make([]byte, capacity)
	binary.BigEndian.PutUint32(bs2, uint32(i))

	dataToSave, err := binary.Append(dataToSave, binary.BigEndian, bs2)
	if err != nil {
		return nil, err
	}

	return dataToSave, nil
}
