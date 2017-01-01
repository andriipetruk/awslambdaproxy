package main

import (
	"os"
	"log"
	"net"
	"io"
	"sync"
)

type LambdaDataCopyManager struct {
	lambdaTunnelConnection *LambdaTunnelConnection
}

func (l *LambdaDataCopyManager) run() {
	for {
		proxySocketConn, proxySocketErr := net.Dial("unix", proxyUnixSocket)
		if proxySocketErr != nil {
			log.Println("Failed to open connection to proxy")
			os.Exit(1)
		}
		log.Println("Started connection to proxy on socket " + proxyUnixSocket)

		tunnelStream, tunnelErr := l.lambdaTunnelConnection.sess.Accept()
		if tunnelErr != nil {
			log.Println("Failed to start stream inside session")
			os.Exit(1)
		}
		log.Println("Started stream inside session")

		go bidirectionalCopy(tunnelStream, proxySocketConn)
	}
}

func newLambdaDataCopyManager(t *LambdaTunnelConnection) *LambdaDataCopyManager {
	return &LambdaDataCopyManager{
		lambdaTunnelConnection: t,
	}
}

func bidirectionalCopy(dst io.ReadWriteCloser, src io.ReadWriteCloser) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		io.Copy(dst, src)
		dst.Close()
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		io.Copy(src, dst)
		src.Close()
		wg.Done()
	}()
}