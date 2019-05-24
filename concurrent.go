package main

import (
	"flag"
	"fmt"
	"math/rand"
	"nanomsg.org/go/mangos/v2/protocol/rep"
	"nanomsg.org/go/mangos/v2/protocol/req"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

import _ "nanomsg.org/go/mangos/v2/transport/all"

const PayloadSize = 600

var sendCount, recvCount uint64

func concurrentFlags() {
	panic("go run concurrent.go [client/server] [benchmark endpoint address]")
}

func main() {
	flag.Parse()

	if len(flag.Args()) < 1 {
		concurrentFlags()
	}

	go func() {
		for range time.Tick(1 * time.Second) {
			fmt.Printf("Sent %d messages, and received %d messages.\n", atomic.SwapUint64(&sendCount, 0), atomic.SwapUint64(&recvCount, 0))
		}
	}()

	switch flag.Arg(0) {
	case "client":
		if len(flag.Args()) != 2 {
			concurrentFlags()
		}

		sock, err := req.NewSocket()
		if err != nil {
			panic(err)
		}

		if err = sock.Dial(flag.Arg(1)); err != nil {
			panic(err)
		}

		go func() {
			msg, err := sock.Recv()
			if err != nil {
				panic(err)
			}

			if len(msg) != PayloadSize {
				panic("Got an unexpected payload size.")
			}

			atomic.AddUint64(&recvCount, 1)
		}()

		for {
			var wg sync.WaitGroup
			wg.Add(16)

			for i := 0; i < 16; i++ {
				go func() {
					var buf [PayloadSize]byte

					if _, err = rand.Read(buf[:]); err != nil {
						panic(err)
					}

					if err = sock.Send(buf[:]); err != nil {
						panic(err)
					}

					atomic.AddUint64(&sendCount, 1)

					wg.Done()
				}()
			}

			wg.Wait()
		}
	case "server":
		if len(flag.Args()) != 1 {
			concurrentFlags()
		}

		sock, err := rep.NewSocket()
		if err != nil {
			panic(err)
		}

		listener, err := sock.NewListener("tcp://:0", nil)
		if err != nil {
			panic(err)
		}

		if err = listener.Listen(); err != nil {
			panic(err)
		}

		addr := listener.Address()

		_, port, err := net.SplitHostPort(addr[strings.Index(addr, "://") + len("://"):])
		if err != nil {
			panic(err)
		}

		fmt.Printf("Listening on port %s.\n", port)

		for {
			msg, err := sock.Recv()
			if err != nil {
				panic(err)
			}

			if len(msg) != 600 {
				panic("Got an unexpected payload size.")
			}

			atomic.AddUint64(&recvCount, 1)

			var buf [PayloadSize]byte

			if _, err = rand.Read(buf[:]); err != nil {
				panic(err)
			}

			if err = sock.Send(buf[:]); err != nil {
				panic(err)
			}

			atomic.AddUint64(&sendCount, 1)
		}
	default:
		concurrentFlags()
	}

}
