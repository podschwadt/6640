package main

import (
	"expvar"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"time"
)

const DefaultTimeout = 1 * time.Second

//const N = 3
//const startPort = 14000

// all messages start with a timestamp as the first "argument"
// a message is a list of strings seperated by a ;
// general format: timestamp;<command>;[parameters]\n
// they must be terminated by a \n

//const REQUEST = "request"       // From
//const INQUIRE = "inquire"       // From
//const RELINQUISH = "relinquish" // From
//const LOCKED = "locked"         // From
//const RELEASE = "rlease"        // From
//const FAILED = "failed"
//
//type Request struct {
//	From        int
//	timestamp int
//}
//
//type State struct {
//	locked          bool
//	request         Request
//	inCS            bool // process is in CS
//	localTime       int  //logical time according to leslie
//	recievedFail    bool
//	recievedInquire bool
//}
//
//func criticalSection(From int) {
//	fmt.Printf("Node %d entering criticalSection\n", From)
//	time.Sleep(time.Duration(rand.Intn(4)+1) * time.Second)
//	fmt.Printf("Node %d leaving criticalSection\n", From)
//}
//
//func create(From int, done chan bool) {
//	port := startPort + From
//	fmt.Printf("creating From: %d \n\t listening on port %d\n", From, port)
//	lisnter, err := net.Listen("tcp", ":"+strconv.Itoa(port))
//	if err != nil {
//		fmt.Print("FUCK")
//	}
//
//	state := State{false, Request{}, false, 0, false, false} // initial state of the node
//
//	quorumSet := buildQuorumSet(From)
//	lockedQueue := make([]int, 0)
//	requestQueue := make([]Request, 0)
//	var mutex = &sync.Mutex{}
//	//server
//	go func() {
//		for {
//			conn, err := lisnter.Accept()
//			if err != nil {
//				fmt.Print("accept failed")
//			}
//			go func(conn net.Conn) {
//				msg, err := bufio.NewReader(conn).ReadString('\n')
//				if err != nil {
//					fmt.Print("too dubm to read")
//				}
//
//				//split message recieved
//				splits := strings.Split(msg, ";")
//
//				// update logical clock
//				senderTimestamp, _ := strconv.Atoi(splits[0])
//				mutex.Lock()
//				if state.localTime+1 < senderTimestamp {
//					state.localTime = senderTimestamp
//				} else {
//					state.localTime++
//				}
//				mutex.Unlock()
//
//				fmt.Printf("Recieved message at node: %d \n \tmessage: %s", From, msg)
//
//				// interpret message
//				cmd := splits[1]
//				senderId, _ := strconv.Atoi(splits[2])
//				switch cmd {
//				case REQUEST:
//					request := Request{senderId, senderTimestamp}
//					if !state.locked {
//						mutex.Lock()
//						state.locked = true
//						state.request = request
//						mutex.Unlock()
//						message := LOCKED + ";" + strconv.Itoa(From)
//						sendMessage(mutex, state, idToPort(senderId), message)
//					} else {
//						mutex.Lock()
//						requestQueue = append(requestQueue, request)
//						mutex.Unlock()
//					}
//					mutex.Lock()
//					// other things go first
//					failed := preceeds(state.request, request)
//					for i := 0; i < len(requestQueue); i++ {
//						if failed || preceeds(requestQueue[i], request) {
//							failed = true
//							break
//						}
//					}
//					mutex.Unlock()
//
//					if failed {
//						message := FAILED + ";" + strconv.Itoa(From)
//						sendMessage(mutex, state, idToPort(senderId), message)
//						return
//					}
//
//					// ask the locking request to chill out
//					sendMessage(mutex, state, idToPort(senderId), INQUIRE)
//				case LOCKED:
//					mutex.Lock()
//					lockedQueue = append(lockedQueue, senderId)
//					mutex.Unlock()
//					//all locks have been recieved
//					if len(lockedQueue) == len(quorumSet) {
//						go func() {
//							criticalSection(From)
//							mutex.Lock()
//							state.inCS = false
//							mutex.Unlock()
//						}()
//					}
//					//empyt the lockedQueue
//					mutex.Lock()
//					lockedQueue = lockedQueue[len(lockedQueue)-1:]
//					mutex.Unlock()
//				case FAILED:
//					if state.recievedInquire {
//						sendMessage(mutex, state, idToPort(senderId), RELINQUISH)
//					}
//					mutex.Lock()
//					state.recievedFail = true
//					mutex.Unlock()
//				case INQUIRE:
//					if state.recievedFail {
//						sendMessage(mutex, state, idToPort(senderId), RELINQUISH)
//					}
//					state.recievedInquire = true
//				case RELINQUISH:
//					mutex.Lock()
//					tempRequest := state.request
//					if len(requestQueue) > 0 {
//						state.request = requestQueue[0]
//						requestQueue[0] = tempRequest
//					} else {
//						state.locked = false
//					}
//					state.recievedFail = false
//					state.recievedInquire = false
//					mutex.Unlock()
//					if state.locked {
//						sendMessage(mutex, state, idToPort(state.request.From), LOCKED)
//					}
//				case RELEASE:
//					mutex.Lock()
//					if len(requestQueue) > 0 {
//						state.request = requestQueue[0]
//						requestQueue = requestQueue[1:]
//					} else {
//						state.locked = false
//					}
//					state.recievedFail = false
//					state.recievedInquire = false
//					mutex.Unlock()
//
//				default:
//					panic(fmt.Sprintf("protocol violation: %s", cmd))
//				}
//
//			}(conn)
//		}
//	}()
//
//	// client
//	// tries to enter the criticalSection
//	go func() {
//		for {
//			delay := rand.Intn(4) + 1
//			fmt.Printf("%d is sleeping for %d \n", From, delay)
//			time.Sleep(time.Duration(delay) * time.Second)
//			for state.locked { /* might as well do nothing  */
//			}
//
//			//locking myself
//			mutex.Lock()
//			state.locked = true
//			state.localTime++
//			state.request = Request{From, state.localTime}
//			mutex.Unlock()
//
//			for i := 0; i < len(quorumSet); i++ {
//				reciverPort := idToPort(quorumSet[i])
//				message := REQUEST + ";" + strconv.Itoa(From)
//				sendMessage(mutex, state, reciverPort, message)
//			}
//			//waiting until we executed the CS before we do request it again
//			mutex.Lock()
//			state.inCS = true
//			mutex.Unlock()
//			for state.inCS {
//				// might as well do nothing
//			}
//		}
//	}()
//
//}
//
//func idToPort(From int) int {
//	return startPort + From
//}
//
//func sendMessage(mutex *sync.Mutex, state State, port int, message string) {
//	conn, err := net.Dial("tcp", "127.0.0.1:"+strconv.Itoa(port))
//	if err != nil {
//		fmt.Print("sent failed")
//	}
//
//	//update logical clock and send message
//	mutex.Lock()
//	state.localTime++
//	fmt.Printf("sending message: %d;%s \n", state.localTime, message)
//	fmt.Fprintf(conn, "%d;%s\n", state.localTime, message)
//	mutex.Unlock()
//}

// func push( queue []int, value int, mutex *sync.Mutex ) {
//   mutex.Lock()
//   *queue = append( *queue, value )
//   mutex.Unlock()
// }
//
// func pop( queue []int, mutex *sync.Mutex ) int {
//   mutex.Lock()
//   defer mutex.Unlock()
//   if len( queue ) == 0{
//     return nil
//   }
//   var value int
//   value, *queue = queue[ 0 ], queue[ 1: ]
//   return value
// }
//
//func buildQuorumSet(From int) []int {
//	// FIXME eventually this should build proper sets
//	quorumSet := make([]int, N-1)
//	tempId := 0
//	for i := 0; i < N-1; {
//		if tempId == From {
//			tempId++
//			continue
//		}
//		quorumSet[i] = tempId
//		tempId++
//		i++
//	}
//	return quorumSet
//}
//
///* does r1 precced r2  */
//func preceeds(r1 Request, r2 Request) bool {
//	return r1.timestamp < r2.timestamp || (r1.timestamp == r2.timestamp && r1.From <= r2.From)
//}

var (
	metrics    = expvar.NewMap("metrics")
	localClock Clock
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	metrics.Init()
}

func main() {
	serverCommand := flag.NewFlagSet("server", flag.ExitOnError)
	serverPortFlag := serverCommand.Int("port", 8000, "Which port to run the orchestrator on.")

	nodeCommand := flag.NewFlagSet("node", flag.ExitOnError)
	nodePortFlag := nodeCommand.Int("port", 10000, "Port to run node on.")
	serverAddressFlag := nodeCommand.String("server", "", "Address of the orchestrator.")

	if len(os.Args) < 2 {
		fmt.Printf("'%v' or '%v' subcommand is required", serverCommand.Name(), nodeCommand.Name())
		os.Exit(1)
	}

	switch os.Args[1] {
	case serverCommand.Name():
		serverCommand.Parse(os.Args[2:])
	case nodeCommand.Name():
		nodeCommand.Parse(os.Args[2:])
	default:
		fmt.Printf("%q is not valid command.\n", os.Args[1])
		os.Exit(2)
	}

	addr := getLocalAddress()
	if serverCommand.Parsed() {
		go startMetrics(addr, *serverPortFlag+1)
		StartOrchestrator(addr, *serverPortFlag)
	} else if nodeCommand.Parsed() {
		go startMetrics(addr, *nodePortFlag+1)
		addr += fmt.Sprintf(":%v", *nodePortFlag)
		self := NewNode(addr)
		self.orchestrator = *serverAddressFlag
		go self.StartNode(*nodePortFlag)
		self.cmdLoop()
	}
}

func startMetrics(addr string, port int) {
	addr += fmt.Sprintf(":%d", port)
	sock, err := net.Listen("tcp", addr)
	if err != nil {
		log.Println("could not start metrics")
	}
	log.Println("Metrics available at:", addr)
	log.Fatal(http.Serve(sock, nil))
}

func getLocalAddress() string {
	conn, err := net.Dial("udp", "1.1.1.1:0")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	addr := conn.LocalAddr().(*net.UDPAddr)
	return addr.IP.String()
}
