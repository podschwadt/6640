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

    locked := false
    waitingForCS := false

    var quorumSet []int
    buildQuorumSet( &quorumSet )

    var lockedQueue []int
    var requestQueue []int
    var mutex = &sync.Mutex{}
    logicalTime := 0 // logical timestamp

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
          if logicalTime + 1 < senderTimestamp {
            logicalTime = senderTimestamp
          } else{
            logicalTime ++
          }
          mutex.Unlock()

          fmt.Printf( "Recieved message at node: %d \n \tmessage: %s", id, msg )

          // interpret message
          cmd := splits[ 1 ]
          switch cmd {
          case REQUEST:
              // id
              senderId, _ := strconv.Atoi( splits[ 2 ] )
              if locked {
                locked = true
                message := LOCKED + ";" + strconv.Itoa( id )
                sendMessage( mutex, logicalTime, idToPort( senderId ), message )
              }else{
                push( &requestQueue, senderId, mutex )
              }
          case LOCKED:
              senderId, _ := strconv.Atoi( splits[ 2 ] )
              push( &lockedQueue, senderId, mutex )
              //all locks have been recieved
              if len( lockedQueue ) == len( quorumSet ){
                go func() {
                  criticalSection( id )
                  mutex.Lock()
                  waitingForCS = false
                  //empyt the lockedQueue
                  lockedQueue = lockedQueue[ len( lockedQueue ) -1 : ]
                  mutex.Unlock()
                }
              }

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
        for locked  {
          // might as well do nothing
        }

        for i := 0; i < len( quorumSet ); i ++ {
          reciver := quorumSet[ i ]
          fmt.Printf( "%d is conntecting ot %d \n", id, startPort + reciver  )
          conn, err := net.Dial( "tcp", "127.0.0.1:" + strconv.Itoa( startPort + reciver ) )
          if err != nil {
            fmt.Print( "sent failed" )
          }

          //update logical clock and send message
          mutex.Lock()
          logicalTime ++
          fmt.Fprintf( conn, "%d;%s;%d\n", logicalTime, REQUEST, id  )
          mutex.Unlock()
        }
        //waiting until we executed the CS before we do request it again
        mutex.Lock()
        waitingForCS = true
        mutex.Unlock()
        for waitingForCS {
          // might as well do nothing
        }
      }
    }()

}

func int idToPort( id int ){
  return startPort + id
}

func sendMessage( mutex *sync.Mutex ,logical_time int, port int, message string){
  conn, err := net.Dial( "tcp", "127.0.0.1:" + strconv.Itoa( startPort + reciver ) )
  if err != nil {
    fmt.Print( "sent failed" )
  }

  //update logical clock and send message
  mutex.Lock()
  logicalTime ++
  fmt.Fprintf( conn, "%d;%s", logicalTime, message  )
  mutex.Unlock()
}

func push( queue *[]int, value int, mutex *sync.Mutex ) {
  mutex.Lock()
  *queue = append( *queue, value )
  mutex.Unlock()
}

func pop( queue *[]int, mutex *sync.Mutex ) {
  mutex.Lock()
  defer mutex.Unlock()
  if len( s ) == 0{
    return nil
  }
  var value int
  value, *queue = queue[ 0 ], queue[ 1: ]
  return value
}

func buildQuorumSet( quorumSet *[]int ) {
  // FIXME eventually this should build proper sets
  *quorumSet = [ N ]int
  for i := 0; i < N; i ++ {
    quorumSet[ i ] = i
  }
}

func main() {
  var done = make( chan bool )
  for i := 0; i < N; i ++ {
    go create( i, done )
  }
  <- done


}
