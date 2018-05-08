package main

import (
	"encoding/gob"
	"log"
	"net"
	"strconv"
	"sync/atomic"
)

type QuorumMessage struct {
	Members []string
}

type InterNodeMessage struct {
	From string
	MsgType
	Clock
}

type MsgType int

const (
	msgRegister MsgType = iota
	msgDeregister

	msgMembershipUpdate

	// Maekawa inter-node communication messages
	msgREQUEST
	msgGRANT
	msgFAILED
	msgINQUIRE
	msgRELEASE
	msgYIELD
)

func NewMsg(from string, t MsgType) *InterNodeMessage {
	localClock.Inc()
	return &InterNodeMessage{
		From:    from,
		MsgType: t,
		Clock:   localClock,
	}
}

func (n *Node) SendMessage(to string, t MsgType) {
	conn, err := net.DialTimeout("tcp", to, DefaultTimeout)
	if err != nil {
		log.Fatalf("Connection error to %v: %v", to, err)
	}
	defer conn.Close()
	encoder := gob.NewEncoder(conn)
	msg := NewMsg(n.Address, t)
	if err := encoder.Encode(msg); err != nil {
		log.Fatalf("Encode message failure: %v", err)
	}
}

func ReceiveMsg(dec *gob.Decoder) *InterNodeMessage {
	msg := &InterNodeMessage{}
	if err := dec.Decode(msg); err != nil {
		return nil
	}
	localClock.Max(msg.Clock)
	localClock.Inc()
	return msg
}

func SendMembership(to string, members []string) {
	conn, err := net.DialTimeout("tcp", to, DefaultTimeout)
	if err != nil {
		log.Println("failed to send membership update to:", to)
		return
	}
	defer conn.Close()
	encoder := gob.NewEncoder(conn)
	msg := NewMsg("", msgMembershipUpdate)
	if err := encoder.Encode(msg); err != nil {
		log.Fatalf("Encode message failure: %v", err)
	}

	qMsg := &QuorumMessage{members}
	if err := encoder.Encode(qMsg); err != nil {
		log.Fatalf("Encode message failure: %v", err)
	}
}

func ReceiveMembership(dec *gob.Decoder) []string {
	msg := &QuorumMessage{}
	if err := dec.Decode(msg); err != nil {
		log.Fatal(err)
		return nil
	}
	return msg.Members
}

// lamport clock wrapper
type Clock uint64

func (c *Clock) String() string { return strconv.FormatUint(uint64(*c), 10) }

func (c *Clock) Inc() {
	atomic.AddUint64((*uint64)(c), 1)
	metrics.Set("clock", c)
}

func (c *Clock) Max(b Clock) {
	if *c < b {
		*c = b
	}
	metrics.Set("clock", c)
}
