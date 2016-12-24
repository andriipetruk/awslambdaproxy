package awslambdaproxy

import (
	"net"
	"sync"
	"log"
	"time"

	"github.com/hashicorp/yamux"
	"github.com/pkg/errors"
	"github.com/Rudd-O/curvetls"
)

type TunnelConnection struct {
	conn net.Conn
	sess *yamux.Session
}

type TunnelConnectionManager struct {
	listener net.Listener
	lastTunnel TunnelConnection
	mutex sync.RWMutex
	serverClientKeypair *ServerClientKeypair
}

func (t *TunnelConnectionManager) run() {
	for {
		c, err := t.listener.Accept()
		if err != nil {
			log.Println("Failed to accept tunnel connection")
			return
		}
		log.Println("Accepted tunnel connection from", c.RemoteAddr())

		long_nonce, err := curvetls.NewLongNonce()
		if err != nil {
			log.Fatalf("Failed to generate nonce: ", err)
		}
		authorizer, clientpubkey, err := curvetls.WrapServer(c, t.serverClientKeypair.serverPrivateKey,
			t.serverClientKeypair.serverPublicKey, long_nonce)
		if err != nil {
			log.Println("Failed to wrap socket: ", err)
			return
		}
		log.Printf("Server: client's public key is %s", clientpubkey)

		var sconn *curvetls.EncryptedConn
		var allowed bool
		if clientpubkey == t.serverClientKeypair.clientPublicKey {
			sconn, err = authorizer.Allow()
			allowed = true
		} else {
			err = authorizer.Deny()
			allowed = false
		}

		if err != nil {
			log.Println("Failed to process authorization: ", err)
			return
		}

		if allowed == true {
			tunnelSession, err := yamux.Client(sconn, nil)
			if err != nil {
				log.Println("Failed to start session inside tunnel")
				return
			}
			log.Println("Started session inside tunnel")

			t.mutex.Lock()
			t.lastTunnel = TunnelConnection{sconn, tunnelSession}
			t.mutex.Unlock()
		}
	}
}

func (t *TunnelConnectionManager) waitUntilReady() {
	for {
		if t.isReady() == true {
			break
		} else {
			log.Println("Waiting for tunnel to be established..")
			time.Sleep(time.Second * 1)
		}
	}
}

func (t *TunnelConnectionManager) isReady() bool {
	if t.lastTunnel == (TunnelConnection{}) {
		return false
	} else {
		return true
	}
}

func newTunnelConnectionManager(port string, serverClientKeyPair *ServerClientKeypair) (*TunnelConnectionManager, error) {
	listener, err := startTunnelListener(port)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to start TunnelConnectionManager")
	}
	return &TunnelConnectionManager{
		listener: listener,
		serverClientKeypair: serverClientKeyPair,
	}, nil
}

func startTunnelListener(tunnelPort string) (net.Listener, error) {
	tunnelAddress := "0.0.0.0:" + tunnelPort
	tunnelListener, err := net.Listen("tcp", tunnelAddress)
	if err != nil {
		errors.Wrap(err, "Failed to start TCP tunnel listener on port " + tunnelPort)
	}
	log.Println("Started TCP tunnel listener on port " + tunnelPort)
	return tunnelListener, nil
}