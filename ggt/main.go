package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	arguments := os.Args
	path, _ := filepath.Abs(".git")
	standardizedPath := filepath.ToSlash(path) + "/"
	fmt.Println()

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
