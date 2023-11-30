package torrent

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/jackpal/bencode-go"
	"github.com/lccatala/JoseLuis/p2p"
	"github.com/lccatala/JoseLuis/peers"
)

const Port uint16 = 8008

type bencodeInfo struct {
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
}

type bencodeTorrent struct {
	Announce string      `bencode:"announce"`
	Info     bencodeInfo `bencode:"info"`
}

type bencodeTrackerResponse struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

// Open parses a torrent file
func Open(filepath string) (TorrentFile, error) {
	f, err := os.Open(filepath)
	if err != nil {
		log.Fatal("Could not read torrent file")
	}
	defer f.Close()

	bto := bencodeTorrent{}
	err = bencode.Unmarshal(f, &bto)
	if err != nil {
		return TorrentFile{}, err
	}

	return bto.toTorrentFile()
}

// DownloadToFile downloads a torrent and writes it to a file
func (tf *TorrentFile) DownloadToFile(path string) error {
	var peerID [20]byte
	_, err := rand.Read(peerID[:])
	if err != nil {
		return err
	}

	peers, err := tf.requestPeers(peerID, Port)
	if err != nil {
		return err
	}

	torrent := p2p.Torrent{
		Peers:       peers,
		PeerID:      peerID,
		InfoHash:    tf.InfoHash,
		PieceHashes: tf.PieceHashes,
		PieceLength: tf.PieceLength,
		Length:      tf.Length,
		Name:        tf.Name,
	}
	buf, err := torrent.Download()
	if err != nil {
		return err
	}

	outFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = outFile.Write(buf)
	if err != nil {
		return err
	}

	return nil
}

func (info *bencodeInfo) hash() ([20]byte, error) {
	var buf bytes.Buffer
	err := bencode.Marshal(&buf, *info)
	if err != nil {
		return [20]byte{}, nil
	}
	h := sha1.Sum(buf.Bytes())
	return h, nil
}

func (info *bencodeInfo) splitPieceHashes() ([][20]byte, error) {
	hashLen := 20
	buf := []byte(info.Pieces)
	if len(buf)%hashLen != 0 {
		err := fmt.Errorf("Received malformed pieces of length %d", len(buf))
		return nil, err
	}

	numHashes := len(buf) / hashLen
	hashes := make([][20]byte, numHashes)

	for i := 0; i < numHashes; i++ {
		copy(hashes[i][:], buf[i*hashLen:(i+1)*hashLen])
	}

	return hashes, nil
}

func (bto *bencodeTorrent) toTorrentFile() (TorrentFile, error) {
	infoHash, err := bto.Info.hash()
	if err != nil {
		return TorrentFile{}, err
	}

	pieceHashes, err := bto.Info.splitPieceHashes()
	if err != nil {
		return TorrentFile{}, err
	}

	torrentFile := TorrentFile{
		Announce:    bto.Announce,
		Name:        bto.Info.Name,
		Length:      bto.Info.Length,
		PieceLength: bto.Info.PieceLength,
		InfoHash:    infoHash,
		PieceHashes: pieceHashes,
	}
	return torrentFile, nil
}

func (t *TorrentFile) buildTrackerURL(peerID [20]byte, port uint16) (string, error) {
	base, err := url.Parse(t.Announce)
	if err != nil {
		return "", err
	}

	params := url.Values{
		"info_hash":  []string{string(t.InfoHash[:])},
		"peer_id":    []string{string(peerID[:])},
		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(t.Length)},
	}

	base.RawQuery = params.Encode()
	return base.String(), nil
}

func (t *TorrentFile) requestPeers(peerID [20]byte, port uint16) ([]peers.Peer, error) {
	url, err := t.buildTrackerURL(peerID, port)
	if err != nil {
		return nil, err
	}

	c := &http.Client{Timeout: 15 * time.Second}
	resp, err := c.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	trackerResp := bencodeTrackerResponse{}
	err = bencode.Unmarshal(resp.Body, &trackerResp)
	if err != nil {
		return nil, err
	}

	return peers.Unmarshal([]byte(trackerResp.Peers))
}
