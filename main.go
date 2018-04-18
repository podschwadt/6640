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

const grant = "grant"
const failed = "failed"
const release = "rlease"
const yield = "yield"



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

    // var grantQueue []int
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
          msg, err := bufio.NewReader(conn).ReadString('\n')
          if err != nil {
            fmt.Print( "to dubm to read" )
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
              senderId, _ := strconv.Atoi( splits[ 1 ] )
              push( &requestQueue, senderId, mutex )
              //strconv.ParseInt( splits[ 2 ] )
          default:
            panic( fmt.Sprintf( "protocol violation: %s", cmd ) )
          }

        }( conn )
      }
    }()

    //client
    go func(){
      for{
        delay := rand.Intn( 4 ) + 1
        fmt.Printf( "%d is sleeping for %d \n", id, delay  )
        time.Sleep( time.Duration( delay ) * time.Second )
        reciver := rand.Intn( N )
        fmt.Printf( "%d is conntecting ot %d \n", id, startPort + reciver  )
        conn, err := net.Dial( "tcp", "127.0.0.1:" + strconv.Itoa( startPort + reciver ) )
        if err != nil {
          fmt.Print( "sent failed" )
        }

        //update logical clock
        mutex.Lock()
        logicalTime ++
        mutex.Unlock()
        fmt.Fprintf( conn, "%d;%s;%d\n", logicalTime, REQUEST, id  )
      }
    }()

}

func push( queue *[]int, value int, mutex *sync.Mutex ) {
  mutex.Lock()
  *queue = append( *queue, value )
  mutex.Unlock()
}

func main() {
  var done = make( chan bool )
  for i := 0; i < N; i ++ {
    go create( i, done )
  }
  <- done


}
