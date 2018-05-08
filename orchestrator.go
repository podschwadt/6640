package main

import (
	"encoding/gob"
	"expvar"
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"sync"
	"time"
)

const healthCheckTick = 1000 * time.Millisecond

var data struct {
	sync.RWMutex
	Nodes map[string]*Node
}

func StartOrchestrator(addr string, port int) {
	data.Nodes = make(map[string]*Node)
	addr += fmt.Sprintf(":%v", port)
	listener, err := net.Listen("tcp", addr)
	defer listener.Close()
	if err != nil {
		log.Fatalf("Socket listen port %d failed,%s", port, err)
		os.Exit(1)
	}
	log.Printf("[=]\tBeginning orchestration at: %v", addr)

	go DoHealthChecks(healthCheckTick)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalln(err)
			continue
		}
		go orchestratorHandler(conn)
	}
}

func orchestratorHandler(conn net.Conn) {
	defer conn.Close()
	dec := gob.NewDecoder(conn)
	msg := ReceiveMsg(dec)

	data.Lock()
	switch msg.MsgType {
	case msgRegister:
		data.Nodes[msg.From] = NewNode(msg.From)
		log.Printf("[+]\tRegistered %v\n", msg.From)
	case msgDeregister:
		delete(data.Nodes, msg.From)
		log.Printf("[-]\tDeregistered %v\n", msg.From)
	}
	data.Unlock()

	UpdateQuorums()
}

func UpdateQuorums() {
	for addr, quorum := range buildQuorum(nodeAddresses()) {
		go SendMembership(addr, quorum)
	}
}

func DoHealthChecks(t time.Duration) {
	tick := time.Tick(t)
	slowTick := time.Tick(t * 3)
	nNodes := &expvar.Int{}
	for {
		select {
		case <-tick:
			for _, node := range data.Nodes {
				go HealthCheck(node, t/3)
			}
		case <-slowTick:
			nNodes.Set(int64(len(nodeAddresses())))
			metrics.Set("nodes", nNodes)
			log.Printf("[%d]\tHealthy nodes", nNodes.Value())
		}
	}
}

func HealthCheck(n *Node, timeout time.Duration) {
	conn, err := net.DialTimeout("tcp", n.Address, timeout)
	if err != nil {
		n.Alive = false
		return
	}
	defer conn.Close()
	n.Alive = true
}

func nodeAddresses() []string {
	data.RLock()
	addrs := make([]string, 0, len(data.Nodes))
	for _, n := range data.Nodes {
		if n.Alive {
			addrs = append(addrs, n.Address)
		}
	}
	data.RUnlock()
	sort.Strings(addrs)
	return addrs
}
