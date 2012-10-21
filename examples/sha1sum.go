package main

import (
	"fmt"
	"io"
	"github.com/Nightgunner5/vpk"
	"os"
	"crypto/sha1"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage:", os.Args[0], "filename.vpk")
		return
	}
	MAIN_FILE := os.Args[1]

	f, err := os.Open(MAIN_FILE)
	eh(err)
	defer f.Close()

	vpkFile, err := vpk.ReadVPKFile(f)
	eh(err)

	for _, filename := range vpkFile.ListFiles() {
		sha := sha1.New()
		data, err := vpkFile.GetReader(vpkFile.GetFileInfo(filename), MAIN_FILE)
		eh(err)
		io.Copy(sha, data)
		eh(data.Close())
		fmt.Printf("%x  %s\n", sha.Sum(nil), filename)
	}
}

func eh(err error) {
	if err != nil {
		panic(err)
	}
}
