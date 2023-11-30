package main

import (
	"log"

	"github.com/lccatala/JoseLuis/torrent"
)

func main() {
	torrentFile, err := torrent.Open("/Users/luis/Downloads/debian-12.2.0-amd64-netinst.iso.torrent")
	if err != nil {
		log.Fatal("Could not load torrent file")
	}

	torrentFile.DownloadToFile("output")
	if err != nil {
		log.Fatal(err)
	}
}
