package main

import (
	"net"
	"os"
	"log"

	"github.com/hashicorp/yamux"
	"github.com/Rudd-O/curvetls"
)

type LambdaTunnelConnection struct {
	tunnelHost string
	conn net.Conn
	sess *yamux.Session
	clientServerKeypair *ClientServerKeypair
}

func (l *LambdaTunnelConnection) setup() {
	tunnelConn, err := net.Dial("tcp", l.tunnelHost)
	if err != nil {
		log.Println("Failed to start tunnel to: ", l.tunnelHost)
		os.Exit(1)
	}
	log.Println("Created tunnel to: " + l.tunnelHost)
	long_nonce, err := curvetls.NewLongNonce()
	if err != nil {
		log.Println("Failed to generate nonce: %s", err)
		return
	}
	sconn, err := curvetls.WrapClient(tunnelConn, l.clientServerKeypair.clientPrivateKey,
			l.clientServerKeypair.clientPublicKey, l.clientServerKeypair.serverPublicKey, long_nonce)
	if err != nil {
		if curvetls.IsAuthenticationError(err) {
			log.Println("Client: server says unauthorized: %s", err)
			return
		} else {
			log.Println("Client: failed to wrap socket: %s", err)
			return
		}
	}
	log.Println("Created tunnel to: " + l.tunnelHost)
	l.conn = sconn


	tunnelSession, err := yamux.Server(l.conn, nil)
	if err != nil {
		log.Println("Failed to start session inside tunnel")
		os.Exit(1)
	}
	log.Println("Started yamux session inside tunnel")
	l.sess = tunnelSession
}

func setupLambdaTunnelConnection(tunnelHost string, clientServerKeypair *ClientServerKeypair) *LambdaTunnelConnection {
	ltc := &LambdaTunnelConnection{
		tunnelHost: tunnelHost,
		clientServerKeypair: clientServerKeypair,
	}
	ltc.setup()
	return ltc
}