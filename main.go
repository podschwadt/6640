package main

import(
  "strconv"
  "fmt"
  "net"
  "time"
  "math/rand"
  "bufio"
  "sync"
  "strings"
)

const N = 5
const startPort = 14000

// all messages start with a timestamp as the first "argument"
// a message is a list of strings seperated by a ;
// general format: timestamp;<command>;[parameters]\n
// they must be terminated by a \n


const REQUEST = "request" // id
const INQUIRE = "inquire" // id
const RELINQUISH = "relinquish" // id
const LOCKED = "locked" // id
const RELEASE = "rlease" // id
const FAILED = "failed"
const YIELD = "yield"


type Request struct {
    nodeId int
    timestamp int
}


type State struct {
  locked bool
  request Request
  inCS bool // process is in CS
  localTime int //logical time according to leslie
}

func criticalSection( id int ){
  fmt.Printf( "Node %d entering criticalSection", id )
  time.Sleep( time.Duration( rand.Intn( 4 ) + 1 ) * time.Second )
  fmt.Printf( "Node %d leaving criticalSection", id )
}


func create( id int, done chan bool ){
    fmt.Printf( "creating id: %d \n", id )
    port := startPort + id
    lisnter, err := net.Listen( "tcp", ":" + strconv.Itoa( port )  )
    if err != nil {
      fmt.Print( "FUCK" )
    }

    state := State{ false, Request{}, false, 0 } // initial state of the node

    quorumSet := buildQuorumSet()
    lockedQueue := make( []int, 0 )
    requestQueue := make( []int, 0 )
    var mutex = &sync.Mutex{}
    //server
    go func(){
      for{
        conn, err := lisnter.Accept()
        if err != nil {
          fmt.Print( "accept failed" )
        }
        go func( conn net.Conn ){
          msg, err := bufio.NewReader( conn ).ReadString( '\n' )
          if err != nil {
            fmt.Print( "too dubm to read" )
          }

          //split message recieved
          splits := strings.Split( msg, ";" )

          // update logical clock
          senderTimestamp, _ := strconv.Atoi( splits[ 0 ] )
          mutex.Lock()
          if state.localTime + 1 < senderTimestamp {
            state.localTime = senderTimestamp
          } else{
            state.localTime ++
          }
          mutex.Unlock()

          fmt.Printf( "Recieved message at node: %d \n \tmessage: %s", id, msg )

          // interpret message
          cmd := splits[ 1 ]
          switch cmd {
          case REQUEST:
              // id
              senderId, _ := strconv.Atoi( splits[ 2 ] )
              if ! state.locked {
                mutex.Lock()
                state.locked = true
                state.request = Request{ senderId, senderTimestamp }
                mutex.Unlock()
                message := LOCKED + ";" + strconv.Itoa( id )
                sendMessage( mutex, state, idToPort( senderId ), message )
              }else{
                mutex.Lock()
                requestQueue = append( requestQueue, senderId )
                mutex.Unlock()
              }
          case LOCKED:
              senderId, _ := strconv.Atoi( splits[ 2 ] )
              mutex.Lock()
              lockedQueue = append( lockedQueue, senderId )
              mutex.Unlock()
              //all locks have been recieved
              if len( lockedQueue ) == len( quorumSet ){
                go func() {
                  criticalSection( id )
                  mutex.Lock()
                  state.inCS = false
                  mutex.Unlock()
                }()
              }
              //empyt the lockedQueue
              mutex.Lock()
              lockedQueue = lockedQueue[ len( lockedQueue ) -1 : ]
              mutex.Unlock()

          default:
            panic( fmt.Sprintf( "protocol violation: %s", cmd ) )
          }

        }( conn )
      }
    }()

    // client
    // tries to enter the criticalSection
    go func(){
      for{
        delay := rand.Intn( 4 ) + 1
        fmt.Printf( "%d is sleeping for %d \n", id, delay  )
        time.Sleep( time.Duration( delay ) * time.Second )
        for state.locked  { /* might as well do nothing  */  }

        for i := 0; i < len( quorumSet ); i ++ {
          reciver := quorumSet[ i ]
          fmt.Printf( "%d is conntecting ot %d \n", id, startPort + reciver  )
          conn, err := net.Dial( "tcp", "127.0.0.1:" + strconv.Itoa( startPort + reciver ) )
          if err != nil { fmt.Print( "sent failed" ) }

          //update logical clock and send message
          mutex.Lock()
          state.localTime ++
          fmt.Fprintf( conn, "%d;%s;%d\n", state.localTime, REQUEST, id  )
          mutex.Unlock()
        }
        //waiting until we executed the CS before we do request it again
        mutex.Lock()
        state.inCS = true
        mutex.Unlock()
        for state.inCS {
          // might as well do nothing
        }
      }
    }()

}


func idToPort( id int ) int {
  return startPort + id
}

func sendMessage( mutex *sync.Mutex, state State, port int, message string){
  fmt.Print( "sending message %s", message )
  conn, err := net.Dial( "tcp", "127.0.0.1:" + strconv.Itoa( port ) )
  if err != nil {
    fmt.Print( "sent failed" )
  }

  //update logical clock and send message
  mutex.Lock()
  state.localTime ++

  fmt.Fprintf( conn, "%d;%s", state.localTime, message  )
  mutex.Unlock()
}

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

func buildQuorumSet() []int {
  // FIXME eventually this should build proper sets
  quorumSet := make( []int, N )
  for i := 0; i < N; i ++ {
    quorumSet[ i ] = i
  }
  return quorumSet
}

func main() {
  var done = make( chan bool )
  for i := 0; i < N; i ++ {
    go create( i, done )
  }
  <- done


}
