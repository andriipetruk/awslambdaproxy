package awslambdaproxy

import (
	"time"
	"os"
	"log"
	"github.com/Rudd-O/curvetls"
)

const (
	lambdaExecutionFrequency = (time.Second * 10)
	lambdaExecutionTimeout = int64(15)
)

type ServerClientKeypair struct {
	serverPrivateKey curvetls.Privkey
	serverPublicKey curvetls.Pubkey
	clientPrivateKey curvetls.Privkey
	clientPublicKey curvetls.Pubkey
}

func newServerClientKeypair() *ServerClientKeypair {
	serverPrivateKey, serverPublicKey, err := curvetls.GenKeyPair()
	if err != nil {
		log.Println("Failed to generate server keypair")
		os.Exit(1)
	}
	log.Println("Server private key: ", serverPrivateKey)
	log.Println("Server public key: ", serverPublicKey)
	clientPrivateKey, clientPublicKey, err := curvetls.GenKeyPair()
	if err != nil {
		log.Println("Failed to generate server keypair")
		os.Exit(1)
	}
	log.Println("Client private key: ", clientPrivateKey)
	log.Println("Client public key: ", clientPublicKey)
	return &ServerClientKeypair{
		serverPrivateKey: serverPrivateKey,
		serverPublicKey: serverPublicKey,
		clientPrivateKey: clientPrivateKey,
		clientPublicKey: clientPublicKey,
	}
}

func ServerInit(proxyPort string, tunnelPort string, regions []string) {
	log.Println("Setting up Lambda infrastructure")
	err := setupLambdaInfrastructure(regions, lambdaExecutionTimeout)
	if err != nil {
		log.Println("Failed to setup Lambda infrastructure", err.Error())
		os.Exit(1)
	}

	serverClientKeypair := newServerClientKeypair()

	log.Println("Starting TunnelConnectionManager")
	tunnelConnectionManager, err := newTunnelConnectionManager(tunnelPort, serverClientKeypair)
	if err != nil {
		log.Println("Failed to setup TunnelConnectionManager", err.Error())
		os.Exit(1)
	}
	go tunnelConnectionManager.run()

	log.Println("Starting LambdaExecutionManager")
	lambdaExecutionManager, err := newLambdaExecutionManager(tunnelPort, regions, lambdaExecutionFrequency, serverClientKeypair)
	if err != nil {
		log.Println("Failed to setup LambdaExecutionManager", err.Error())
		os.Exit(1)
	}
	go lambdaExecutionManager.run()

	tunnelConnectionManager.waitUntilReady()

	log.Println("Starting UserConnectionManager")
	userConnectionManager, err := newUserConnectionManager(proxyPort)
	if err != nil {
		log.Println("Failed to setup UserConnectionManager", err.Error())
		os.Exit(1)
	}
	go userConnectionManager.run()

	log.Println("Starting DataCopyManager")
	dataCopyManager := newDataCopyManager(userConnectionManager, tunnelConnectionManager)
	dataCopyManager.run()
}