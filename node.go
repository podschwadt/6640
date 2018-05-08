package main

import (
	"bufio"
	"encoding/gob"
	"expvar"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	currPrompt       string
	infoPrompt       = func(i string) { fmt.Printf("\n[%d] %v\n%v", localClock, i, currPrompt) }
	cmdNotRecognized = func(c string) { fmt.Printf("! Command not recognized: '%v'\n", c) }
)

type Node struct {
	Address      string
	Alive        bool
	Quorum       []string // other quorum addresses
	grantMap     map[string]bool
	reqQ         *RequestQ
	orchestrator string
	registeredAt *time.Time
}

func (n Node) String() string { return fmt.Sprintf("[%v]{UP: %v}", n.Address, n.Alive) }

func NewNode(addr string) *Node {
	return &Node{
		Address: addr,
		Alive:   true,
	}
}

func (n *Node) StartNode(port int) {
	listener, err := net.Listen("tcp", n.Address)
	if err != nil {
		log.Fatalf("could not establish listener: %v", err)
	}
	log.Printf("[%v] Started listening\n", n.Address)
	defer listener.Close()

	n.resetGrantMap()
	n.reqQ = NewRequestQ()
	n.Register()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalln(err)
			continue
		}
		go n.nodeHandler(conn)
	}
}

func (n *Node) nodeHandler(conn net.Conn) {
	defer conn.Close()
	dec := gob.NewDecoder(conn)
	msg := ReceiveMsg(dec)
	if msg == nil {
		return
	}

	from := msg.From
	switch msg.MsgType {
	case msgMembershipUpdate:
		oldMembers := n.Quorum
		n.Quorum = ReceiveMembership(dec)
		metrics.Set("quorum", expvar.Func(n.quorum))
		infoPrompt(fmt.Sprintf("Received new members: %v => %v\n", oldMembers, n.Quorum))
		n.resetGrantMap()
	case msgREQUEST:
		if n.reqQ.Len() == 0 {
			n.SendMessage(msg.From, msgGRANT)
			n.reqQ.Push(NewRequest(from))
			infoPrompt(fmt.Sprintf("Received Request from %v and sent Grant to %v", from, from))
		} else if Less(n.reqQ.Peek().(Request), *msg) {
			n.SendMessage(msg.From, msgFAILED)
			n.reqQ.Push(NewRequest(from))
			infoPrompt(fmt.Sprintf("Received Request from %v and sent Failed to %v", from, from))
		} else {
			request := n.reqQ.Peek().(Request)
			n.SendMessage(request.From, msgINQUIRE)
			infoPrompt(fmt.Sprintf("Received Request from %v and sent Inquire to %v", from, request.From))
		}
	case msgGRANT:
		n.grantMap[from] = true
		infoPrompt(fmt.Sprintf("Received Grant from %v", from))
		if n.LockedQuorum() {
			timeInCS := 2 + rand.Intn(3)
			done := time.NewTimer(time.Duration(timeInCS) * time.Second).C
			fmt.Printf("\nHave received all Grants => starting Critical Section for %ds\n", timeInCS)
			ticker := time.NewTicker(500 * time.Millisecond)
		CritialSection:
			for {
				select {
				case <-done:
					fmt.Println("Complete!")
					break CritialSection
				case <-ticker.C:
					fmt.Printf(" Work ")
				}
			}
			n.SendReleases()
			infoPrompt("All work is done => Exiting Critical Section")
		}
	case msgFAILED:
		qLen := len(*n.reqQ) - 1
		if (*n.reqQ)[qLen].inquire == nil {
			(*n.reqQ)[qLen].failed = true
			infoPrompt(fmt.Sprintf("Received Failed from %v", from))
			return
		}
		inquire := (*n.reqQ)[qLen].inquire
		n.reqQ.Swap(qLen, qLen-1)
		n.SendMessage(inquire.From, msgYIELD)
		infoPrompt(fmt.Sprintf("Received Failed from %v and sent Yield to %v", from, inquire.From))
	case msgINQUIRE:
		qLen := len(*n.reqQ) - 1
		if (*n.reqQ)[qLen].failed {
			(*n.reqQ)[qLen].failed = false
			n.SendMessage(from, msgYIELD)
			infoPrompt(fmt.Sprintf("Received Inquire from %v and sent Yield to %v", from, from))
			n.reqQ.Swap(qLen, qLen-1)
			return
		}
		(*n.reqQ)[qLen].inquire = msg
	case msgYIELD:
		request := n.reqQ.Peek().(Request)
		infoPrompt(fmt.Sprintf("Received Yield from %v and sent Grant to %v", from, request.From))
		n.SendMessage(request.From, msgGRANT)
		n.reqQ.Push(NewRequest(from))
		n.resetGrantMap()
	case msgRELEASE:
		n.reqQ.Pop() // release current grant
		if n.reqQ.Len() > 0 {
			if request, ok := n.reqQ.Pop().(Request); ok {
				n.SendMessage(request.From, msgRELEASE)
			}
		}
		n.resetGrantMap()
		infoPrompt(fmt.Sprintf("Received Release from %v (queued: %v)", from, n.reqQ.Len()))
	}
}

func (n *Node) cmdLoop() {
	time.Sleep(250 * time.Millisecond)
	for {
		n.unregisteredLoop()
		n.registeredLoop()
	}
}

func (n *Node) unregisteredLoop() {
	currPrompt = fmt.Sprint(" [r] Register\t[q] Quit\n: ")
	for n.registeredAt == nil {
		cmd := getInput()
		switch cmd {
		case "r":
			n.Register()
		case "q":
			os.Exit(0)
		default:
			cmdNotRecognized(cmd)
		}
	}
}

func (n *Node) registeredLoop() {
	currPrompt = fmt.Sprint(" [r] Request Resource\t[d] Deregister\t[q] Quit\n: ")
	for n.registeredAt != nil {
		cmd := getInput()
		switch cmd {
		case "r":
			n.reqQ.Push(NewRequest(n.Address))
			n.SendRequests()
			n.waitingForLockLoop()
		case "d":
			n.Deregister()
		case "q":
			os.Exit(0)
		default:
			cmdNotRecognized(cmd)
		}
	}
}

func (n *Node) waitingForLockLoop() {
	for !n.LockedQuorum() {
		fmt.Print("\r Waiting for lock")
	}
}

func getInput() string {
	fmt.Printf("[%d]", localClock)
	fmt.Print(currPrompt)
	reader := bufio.NewReader(os.Stdin)
	s, _ := reader.ReadString('\n')
	s = strings.TrimSpace(s)
	return s
}

func (n *Node) Register() {
	n.SendMessage(n.orchestrator, msgRegister)
	now := time.Now()
	n.registeredAt = &now
}

func (n *Node) Deregister() {
	n.SendMessage(n.orchestrator, msgDeregister)
	n.registeredAt = nil
}

func (n *Node) SendRequests() {
	n.grantMap[n.Address] = true
	var wg sync.WaitGroup
	wg.Add(len(n.Quorum))
	for _, addr := range n.Quorum {
		go func(a string) {
			n.SendMessage(a, msgREQUEST)
			wg.Done()
		}(addr)
	}
	wg.Wait()
}

func (n *Node) SendReleases() {
	var wg sync.WaitGroup
	wg.Add(len(n.Quorum))
	for _, addr := range n.Quorum {
		go func(a string) {
			n.SendMessage(a, msgRELEASE)
			wg.Done()
		}(addr)
	}
	wg.Wait()
}

func (n *Node) quorum() interface{} {
	return n.Quorum
}

func (n *Node) LockedQuorum() bool {
	for _, g := range n.grantMap {
		if !g {
			return false
		}
	}
	return true
}

func (n *Node) resetGrantMap() {
	n.grantMap = make(map[string]bool, len(n.Quorum)+1)
	n.grantMap[n.Address] = false
	for _, node := range n.Quorum {
		n.grantMap[node] = false
	}
}
