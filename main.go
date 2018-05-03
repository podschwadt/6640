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

const N = 3
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



type Request struct {
    id int
    timestamp int
}


type State struct {
  locked bool
  request Request
  inCS bool // process is in CS
  localTime int //logical time according to leslie
  recievedFail bool
  recievedInquire bool
}

func criticalSection( id int ){
  fmt.Printf( "Node %d entering criticalSection\n", id )
  time.Sleep( time.Duration( rand.Intn( 4 ) + 1 ) * time.Second )
  fmt.Printf( "Node %d leaving criticalSection\n", id )
}


func create( id int, done chan bool ){
    port := startPort + id
    fmt.Printf( "creating id: %d \n\t listening on port %d\n", id, port )
    lisnter, err := net.Listen( "tcp", ":" + strconv.Itoa( port )  )
    if err != nil {
      fmt.Print( "FUCK" )
    }

    state := State{ false, Request{}, false, 0, false, false } // initial state of the node

    quorumSet := buildQuorumSet( id )
    lockedQueue := make( []int, 0 )
    requestQueue := make( []Request, 0 )
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
          senderId, _ := strconv.Atoi( splits[ 2 ] )
          switch cmd {
          case REQUEST:
              request := Request{ senderId, senderTimestamp }
              if ! state.locked {
                mutex.Lock()
                state.locked = true
                state.request = request
                mutex.Unlock()
                message := LOCKED + ";" + strconv.Itoa( id )
                sendMessage( mutex, state, idToPort( senderId ), message )
              } else {
                mutex.Lock()
                requestQueue = append( requestQueue, request )
                mutex.Unlock()
              }
              mutex.Lock()
              // other things go first
              failed := preceeds( state.request, request )
              for i:= 0; i < len( requestQueue ); i++ {
                  if failed || preceeds( requestQueue[ i ], request ){
                    failed = true
                    break
                  }
              }
                mutex.Unlock()

                if failed {
                  message := FAILED + ";" + strconv.Itoa( id )
                  sendMessage( mutex, state, idToPort( senderId ), message )
                  return
                }

                // ask the locking request to chill out
                sendMessage( mutex, state, idToPort( senderId ), INQUIRE )
          case LOCKED:
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
          case FAILED:
              if state.recievedInquire {
                sendMessage( mutex, state, idToPort( senderId ), RELINQUISH )
              }
              mutex.Lock()
              state.recievedFail = true
              mutex.Unlock()
          case INQUIRE:
              if state.recievedFail {
                sendMessage( mutex, state, idToPort( senderId ), RELINQUISH )
              }
              state.recievedInquire = true
          case RELINQUISH:
              mutex.Lock()
              tempRequest := state.request
              if len( requestQueue ) > 0 {
                  state.request = requestQueue[ 0 ]
                  requestQueue[ 0 ] = tempRequest
              } else {
                  state.locked = false
              }
              state.recievedFail = false
              state.recievedInquire = false
              mutex.Unlock()
              if state.locked {
                  sendMessage( mutex, state, idToPort( state.request.id ), LOCKED )
              }
          case RELEASE:
              mutex.Lock()
              if len( requestQueue ) > 0 {
                  state.request = requestQueue[ 0 ]
                  requestQueue = requestQueue [ 1: ]
              } else {
                  state.locked = false
              }
              state.recievedFail = false
              state.recievedInquire = false
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
      for {
        delay := rand.Intn( 4 ) + 1
        fmt.Printf( "%d is sleeping for %d \n", id, delay  )
        time.Sleep( time.Duration( delay ) * time.Second )
        for state.locked  { /* might as well do nothing  */  }

        //locking myself
        mutex.Lock()
        state.locked = true
        state.localTime ++
        state.request = Request{ id, state.localTime }
        mutex.Unlock()

        for i := 0; i < len( quorumSet ); i ++ {
          reciverPort := idToPort( quorumSet[ i ] )
          message := REQUEST + ";" + strconv.Itoa( id )
          sendMessage( mutex, state, reciverPort, message )
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
  conn, err := net.Dial( "tcp", "127.0.0.1:" + strconv.Itoa( port ) )
  if err != nil {
    fmt.Print( "sent failed" )
  }

  //update logical clock and send message
  mutex.Lock()
  state.localTime ++
  fmt.Printf( "sending message: %d;%s \n", state.localTime, message )
  fmt.Fprintf( conn, "%d;%s\n", state.localTime, message  )
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

func buildQuorumSet( id int ) []int {
  // FIXME eventually this should build proper sets
  quorumSet := make( []int, N -1 )
  tempId := 0
  for i := 0; i < N - 1; {
    if tempId == id {
      tempId ++
      continue
    }
    quorumSet[ i ] = tempId
    tempId ++
    i ++
  }
  return quorumSet
}

/* does r1 precced r2  */
func preceeds( r1 Request, r2 Request ) bool {
  return r1.timestamp < r2.timestamp || ( r1.timestamp == r2.timestamp && r1.id <= r2.id )
}

func main() {
  var done = make( chan bool )
  for i := 0; i < N; i ++ {
    go create( i, done )
  }
  <- done


}
