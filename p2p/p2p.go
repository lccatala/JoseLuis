package main

import (
	"github.com/lccatala/JoseLuis/client"
	"github.com/lccatala/JoseLuis/message"
	"github.com/lccatala/JoseLuis/peers"
)

type Torrent struct {
	Peers []peers.Peer
}
