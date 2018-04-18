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

const request = "request"
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


    //server
    go func(){
      for{
        conn, err := lisnter.Accept()
        if err != nil {
          fmt.Print( "accept failed" )
        }
        go func(){
          conn := conn
          msg, err := bufio.NewReader(conn).ReadString('\n')
          if err != nil {
            fmt.Print( "to dubm to read" )
          }
          fmt.Printf( "Recieved message at node: %d \n \tmessage: %s", id, msg )

          splits := strings.Split( msg, ";" )

          switch splits[ 0 ] {
          case request:
              // id, timestamp
              senderId, _ := strconv.Atoi( splits[ 1 ] )
              push( &requestQueue, senderId, mutex )
              //strconv.ParseInt( splits[ 2 ] )
          default:
            panic( "protocol violation" )
          }

        }()
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
        fmt.Fprintf( conn, "request;%d", id  )
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
