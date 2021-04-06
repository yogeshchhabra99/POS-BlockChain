package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	mrand "math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/joho/godotenv"
	golog "github.com/ipfs/go-log"
	libp2p "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
	net "github.com/libp2p/go-libp2p-core/network"
	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
)

type TransactionType string 
const(
	DEBIT="DEBIT"
	CREDIT="CREDIT"
	NEW_ACCOUNT="NEW_ACCOUNT"
	NONE="NONE"
)

type Color string

const (
    ColorBlack  Color = "\u001b[30m"
    ColorRed          = "\u001b[31m"
    ColorGreen        = "\u001b[32m"
    ColorYellow       = "\u001b[33m"
    ColorBlue         = "\u001b[34m"
    ColorReset        = "\u001b[0m"
)

type Block struct {
	Index     int
	Timestamp string
	UserId       int
	Balance		int
	Type		TransactionType
	Amount		int
	Hash      string
	PrevHash  string
}

var Blockchain []Block

var mutex = &sync.Mutex{}
//var isCash *bool
//var isRetail *bool

func calculateHash(block Block) string {
	record := string(block.Index) + block.Timestamp + string(block.UserId) + string(block.Balance) + string(block.Type) + block.PrevHash
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

func makeBasicHost(listenPort int, secio bool, randseed int64) (host.Host, error) {

	// If the seed is zero, use real cryptographic randomness. Otherwise, use a
	// deterministic randomness source to make generated keys stay the same
	// across multiple runs
	var r io.Reader
	if randseed == 0 {
		r = rand.Reader
	} else {
		r = mrand.New(mrand.NewSource(randseed))
	}

	// Generate a key pair for this host. We will use it
	// to obtain a valid host ID.
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		return nil, err
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", listenPort)),
		libp2p.Identity(priv),
	}

	// if !secio {
	// 	opts = append(opts, libp2p.NoEncryption())
	// }

	basicHost, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	// Build host multiaddress
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", basicHost.ID().Pretty()))

	// Now we can build a full multiaddress to reach this host
	// by encapsulating both addresses:
	addr := basicHost.Addrs()[0]
	fullAddr := addr.Encapsulate(hostAddr)
	log.Printf("I am %s\n", fullAddr)
	if secio {
		log.Printf("Now run \"go run p2p_blockchain.go -l %d -d %s -secio\" on a different terminal.\n", listenPort+1, fullAddr)
	} else {
		log.Printf("Now run \"go run p2p_blockchain.go -l %d -d %s\" on a different terminal.\n", listenPort+1, fullAddr)
	}

	return basicHost, nil
}

func handleStream(s net.Stream) {

	colorize(ColorBlue,"HANDLESTREAM")

	log.Println("Got a new stream!")

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	go readData(rw)
	go writeData(rw)

	// stream 's' will stay open until you close it (or the other side closes it).
}

func readData(rw *bufio.ReadWriter) {

	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		if str == "" {
			return
		}
		if str != "\n" {

			chain := make([]Block, 0)
			if err := json.Unmarshal([]byte(str), &chain); err != nil {
				log.Fatal(err)
			}

			mutex.Lock()
			if len(chain) > len(Blockchain) {
				Blockchain = chain
				bytes, err := json.MarshalIndent(Blockchain, "", "  ")
				if err != nil {

					log.Fatal(err)
				}
				// Green console color: 	\x1b[32m
				// Reset console color: 	\x1b[0m
				fmt.Printf("\x1b[32m%s\x1b[0m> ", string(bytes))
			}
			mutex.Unlock()
		}
	}
}

func writeData(rw *bufio.ReadWriter) {

	isCash:=true
	firstcommand:=true
	go func() {
		for {
			time.Sleep(5 * time.Second)
			mutex.Lock()
			bytes, err := json.Marshal(Blockchain)
			if err != nil {
				log.Println(err)
			}
			mutex.Unlock()

			mutex.Lock()
			rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
			rw.Flush()
			mutex.Unlock()

		}
	}()

	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")

		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		sendData = strings.Replace(sendData, "\n", "", -1)

		stringArray := strings.Fields(sendData)

		//colorize(ColorBlue,"Debug1")
		if firstcommand==true {
			if stringArray[0]!="CASH" && stringArray[0]!="RETAIL" {
				colorize(ColorRed,"Invalid command the first command has to be either 1. CASH or 2.) RETAIL . Try Again")
			} else {
				firstcommand=false
				if stringArray[0]=="CASH" {
					isCash=true
					colorize(ColorBlue,"This became a Cash POS Terminal")
				} else{
					isCash=false
					colorize(ColorBlue,"This became a Retail POS Terminal")
				}
			}

		} else if isCash==true {
			//colorize(ColorBlue,"Debug2")

			if stringArray[0]!= "NEWUSER" && stringArray[0]!="RECHARGE" {
				colorize(ColorRed,"Invalid Command. This is Cash POS. Valid commands are: 1.) NEWUSER 2.) RECHARGE $USERID $AMOUNT\n")
			}
			if stringArray[0]=="NEWUSER" {
				numUsers := 0
				for i :=0; i<len(Blockchain); i++ {
					if Blockchain[i].UserId>numUsers {
						numUsers = Blockchain[i].UserId
					}
				}
		
				userId:=numUsers+1
				newBlock, err := generateBlock(Blockchain[len(Blockchain)-1], userId,0,NEW_ACCOUNT,0)
				colorize(ColorYellow,"New User account with UserId="+strconv.Itoa(userId)+" is created")

				if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {
					mutex.Lock()
					Blockchain = append(Blockchain, newBlock)
					mutex.Unlock()
				}

				bytes, err := json.Marshal(Blockchain)
				if err != nil {
					log.Println(err)
				}

				spew.Dump(Blockchain)

				mutex.Lock()
				rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
				rw.Flush()
				mutex.Unlock()
			}
			if stringArray[0]=="RECHARGE" {
				colorize(ColorBlue,"REcharge\n")
				userId, err := strconv.Atoi(stringArray[1])
				if err !=nil {
					colorize(ColorRed,"userID must be a number\n")
					log.Fatal(err)
				} else {
					userFound:=false
					balance:=0
					for i :=len(Blockchain)-1; i>=0; i-- {
						if Blockchain[i].UserId==userId {
							userFound = true
							balance= Blockchain[i].Balance
							break
						}
					}
					if userFound==true {
						amount, err := strconv.Atoi(stringArray[2])
						if err !=nil {
							colorize(ColorRed,"Amount must be a number\n")
							log.Fatal(err)
						} else if amount <0 {
							colorize(ColorRed,"Amount must be a positive number\n")
						} else{
							balance=balance+amount
							newBlock, err := generateBlock(Blockchain[len(Blockchain)-1], userId,balance,CREDIT,amount)
							colorize(ColorYellow,"Amount Credited")

							if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {
								mutex.Lock()
								Blockchain = append(Blockchain, newBlock)
								mutex.Unlock()
							}

							bytes, err := json.Marshal(Blockchain)
							if err != nil {
								log.Println(err)
							}

							spew.Dump(Blockchain)

							mutex.Lock()
							rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
							rw.Flush()
							mutex.Unlock()
						}
					} else {
						colorize(ColorRed,"Account with given userID not found\n")
					}
				}
			}
		} else {
			//Retail type POS
			if stringArray[0]!= "DEBIT"  {
				colorize(ColorRed,"Invalid Command. This is Retail POS. Valid commands are: 1.) DEBIT $USERID $AMOUNT\n")
			} else{
				userId, err := strconv.Atoi(stringArray[1])
				if err !=nil {
					colorize(ColorRed,"userID must be a number\n")
					log.Fatal(err)
				} else {
					userFound:=false
					balance:=0
					for i :=len(Blockchain)-1; i>=0; i-- {
						if Blockchain[i].UserId==userId {
							userFound = true
							balance= Blockchain[i].Balance
							break
						}
					}
					if userFound==true {
						amount, err := strconv.Atoi(stringArray[2])
						if err !=nil {
							colorize(ColorRed,"Amount must be a number\n")
							log.Fatal(err)
						} else if amount <0 {
							colorize(ColorRed,"Amount must be a positive number\n")
						} else if balance<amount{
							colorize(ColorRed,"Balance less than debit amount\n")
						} else{
							balance:=balance-amount
							newBlock, err := generateBlock(Blockchain[len(Blockchain)-1], userId,balance,DEBIT,amount)
							colorize(ColorYellow,"Amount Debitted")

							if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {
								mutex.Lock()
								Blockchain = append(Blockchain, newBlock)
								mutex.Unlock()
							}

							bytes, err := json.Marshal(Blockchain)
							if err != nil {
								log.Println(err)
							}

							spew.Dump(Blockchain)

							mutex.Lock()
							rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
							rw.Flush()
							mutex.Unlock()
						}
					} else {
						colorize(ColorRed,"Account with given userID not found\n")
					}
				}
			}
		}
	}

}

func generateBlock(oldBlock Block, userId int, balance int, type_ TransactionType, amount int ) (Block, error) {

	var newBlock Block

	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.UserId=userId
	newBlock.Balance=balance
	newBlock.Type=type_
	newBlock.Amount=amount
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = calculateHash(newBlock)

	return newBlock, nil
}

func isBlockValid(newBlock, oldBlock Block) bool {
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}

	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}

	if calculateHash(newBlock) != newBlock.Hash {
		return false
	}

	return true
}

func replaceChain(newBlocks []Block) {
	if len(newBlocks) > len(Blockchain) {
		Blockchain = newBlocks
	}
}

func colorize(color Color, message string) {
    fmt.Println(string(color), message, string(ColorReset))
}

func main() {

	// Parse options from the command line
	listenF := flag.Int("l", 0, "wait for incoming connections")
	target := flag.String("d", "", "target peer to dial")
	secio := flag.Bool("secio", false, "enable secio")
	seed := flag.Int64("seed", 0, "set random seed for id generation")
	//isCash :=flag.Bool("CASH",false,"Is the POS cash type")
	//isRetail :=flag.Bool("RETAIL",false,"Is the POS retail type")
	flag.Parse()

	// if *isCash==false && *isRetail == false {
	// 	log.Fatal("Please provide either -CASH or -RETAIL flag")
	// 	return
	// }

	// if *isCash==true && *isRetail == true {
	// 	log.Fatal("Please pass only one -CASH or -RETAIL flag")
	// 	return
	// }

	// if *isCash==true {
	// 	colorize(ColorBlue, "This is a Cash POS terminal")
	// }

	// if *isRetail==true {
	// 	colorize(ColorBlue, "This is a Retail POS terminal")
	// }

	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	// create genesis block
	t := time.Now()
	genesisBlock := Block{0, t.String(), 0, 0, NONE, 0, "", ""}
	spew.Dump(genesisBlock)
	Blockchain = append(Blockchain, genesisBlock)

	lvl, err := golog.LevelFromString("info")
	if err==nil {
		golog.SetAllLoggers(lvl) // Change to DEBUG for extra info
	}

	if *listenF == 0 {
		log.Fatal("Please provide a port to bind on with -l")
	}

	// Make a host that listens on the given multiaddress
	ha, err := makeBasicHost(*listenF, *secio, *seed)
	if err != nil {
		log.Fatal(err)
	}

	if *target == "" {
		log.Println("listening for connections")
		// Set a stream handler on host A. /p2p/1.0.0 is
		// a user-defined protocol name.
		ha.SetStreamHandler("/p2p/1.0.0", handleStream)

		select {} // hang forever
		/**** This is where the listener code ends ****/
	} else {
		ha.SetStreamHandler("/p2p/1.0.0", handleStream)

		// The following code extracts target's peer ID from the
		// given multiaddress
		ipfsaddr, err := ma.NewMultiaddr(*target)
		if err != nil {
			log.Fatalln(err)
		}

		pid, err := ipfsaddr.ValueForProtocol(ma.P_IPFS)
		if err != nil {
			log.Fatalln(err)
		}

		peerid, err := peer.IDB58Decode(pid)
		if err != nil {
			log.Fatalln(err)
		}

		// Decapsulate the /ipfs/<peerID> part from the target
		// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>
		targetPeerAddr, _ := ma.NewMultiaddr(
			fmt.Sprintf("/ipfs/%s", peer.IDB58Encode(peerid)))
		targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)

		// We have a peer ID and a targetAddr so we add it to the peerstore
		// so LibP2P knows how to contact it
		ha.Peerstore().AddAddr(peerid, targetAddr, pstore.PermanentAddrTTL)

		log.Println("opening stream")
		// make a new stream from host B to host A
		// it should be handled on host A by the handler we set above because
		// we use the same /p2p/1.0.0 protocol
		s, err := ha.NewStream(context.Background(), peerid, "/p2p/1.0.0")
		if err != nil {
			log.Fatalln(err)
		}
		// Create a buffered stream so that read and writes are non blocking.
		rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

		// Create a thread to read and write data.
		go writeData(rw)
		go readData(rw)

		select {} // hang forever

	}
}


