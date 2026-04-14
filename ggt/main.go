package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

func main() {
	arguments := os.Args
	path, _ := filepath.Abs(".git")
	standardizedPath := filepath.ToSlash(path) + "/"
	fmt.Println()

	// hash-object --stdin
	if len(arguments) == 3 && arguments[1] == "hash-object" && arguments[2] == "--stdin" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Printf("failed to read data: %v", err)
			return
		}

		size := len(data)
		header := fmt.Sprintf("blob %s%c", strconv.Itoa(size), 0)
		content := header + string(data)
		hash := sha1.New()
		hash.Write([]byte(content))
		hashedData := hash.Sum(nil)
		hashedString := hex.EncodeToString(hashedData)
		fmt.Println(hashedString)
		os.Exit(0)
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
