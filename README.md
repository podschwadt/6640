# Maekawa on a budget

DOCU: https://github.com/podschwadt/6640/wiki

## Installation

```bash
go get github.com/podschwadt/6640
cd $GOPATH/src/github.com/podschwadt/6640
go build
```

## Running

### Server

```bash
$ ./6640 server -port 9000
```

### Nodes

```bash
$ ./6640 node -server $SERVER_ADDR -port 10000
```
