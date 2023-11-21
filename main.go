package main

import (
	"fmt"
	"os"
)

func main() {
	filename := "/Users/luis/Downloads/debian-12.2.0-amd64-netinst.iso.torrent"
	f, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Cannot read file %q", filename)
		os.Exit(-1)
	}

	torrent, err := Open(f)
	torrentFile, err := torrent.toTorrentFile()
	if err != nil {
		fmt.Printf("Cannot convert to torrent file")
		os.Exit(-1)
	}
	fmt.Printf("Announce is: %q\n", torrentFile.Announce)

}
