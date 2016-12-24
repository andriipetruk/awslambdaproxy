package main

import (
	"log"
	"os"
	"flag"
	"github.com/Rudd-O/curvetls"
)

const (
	proxyUnixSocket = "/tmp/lambda-proxy.socket"
)

type ClientServerKeypair struct {
	serverPublicKey curvetls.Pubkey
	clientPrivateKey curvetls.Privkey
	clientPublicKey curvetls.Pubkey
}

func newClientServerKeypair(clientPrivateKey string, clientPublicKey string, serverPublicKey string) *ClientServerKeypair {
	cpr, _ := curvetls.PrivkeyFromString(clientPrivateKey)
	cpu, _ := curvetls.PubkeyFromString(clientPublicKey)
	spu, _ := curvetls.PubkeyFromString(serverPublicKey)
	return &ClientServerKeypair{
		serverPublicKey: spu,
		clientPrivateKey: cpr,
		clientPublicKey: cpu,
	}
}

func LambdaInit(tunnelHost string, clientPrivateKey string, clientPublicKey string, serverPublicKey string) {
	log.Println("Starting LambdaProxyServer")
	lambdaProxyServer := startLambdaProxyServer()

	clientServerKeypair := newClientServerKeypair(clientPrivateKey, clientPublicKey, serverPublicKey)

	log.Println("Establishing tunnel connection to", tunnelHost)
	lambdaTunnelConnection := setupLambdaTunnelConnection(tunnelHost, clientServerKeypair)

	log.Println("Starting LambdaDataCopyManager")
	dataCopyManager := newLambdaDataCopyManager(lambdaProxyServer, lambdaTunnelConnection)
	dataCopyManager.run()
}

func main() {
	addressPtr := flag.String("address", "localhost:8081", "IP and port of server to connect to")
	clientPrivateKeyPtr := flag.String("client-private-key", "", "Client private key")
	clientPublicKeyPtr := flag.String("client-public-key", "", "Client public key")
	serverPublicKeyPtr := flag.String("server-public-key", "", "Server public key")

	flag.Parse()

	if *addressPtr == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	if *clientPrivateKeyPtr == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	if *clientPublicKeyPtr == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	if *serverPublicKeyPtr == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}


	LambdaInit(*addressPtr, *clientPrivateKeyPtr, *clientPublicKeyPtr, *serverPublicKeyPtr)

}

