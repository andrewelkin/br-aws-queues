
### What is this?

This is an implementation of client/server interacting with AWS [SQS](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/welcome.html) (simple queue service)
The server receives <key><payload> entries from one or more clients through the SQS queue.
The server keeps the order of entries it receives.   

### Files in the [repository](https://github.com/andrewelkin/br-aws-queues)


```
.
├── aws-helper
│   └── aws-helper.go        -- aws sqs functionality
├── client
│   └── client.go            -- client main()
├── go.mod
├── go.sum
├── processor
│   ├── ordered-map.go       -- ordered map   
│   ├── ordered-map_test.go  -- unit tests
│   └── server-queue.go      -- server queue (based on the ordered map)
├── readme.md                      -- this file
└── server
    └── server.go                  -- server main()

```

### Server internal queue operations

Server operates on its internal queue and understands these types of requests 
from the clients:

| Op code | Function | Description   |
|---------|----------|---------------|
| '+'     | AddItem  | Add item      |
| '-'     | DelItem  | Delete item   |
| '?'     | GetItem  | Request item  |
| '*'     | GetAll   | Get all items |




### How the server works

The server creates a SQS queue (command line parameter 'qname', default name is 'bl-test-queue')) and
awaits for messages. When one or more messages arrive, the server 
calls interface method 
```ProcPayload(string)``` of the initialized payload processor.
cmd line arg 'threads' defines number of threads in the payload processor (default is 3).

Multithreaded payload processor creates a pool of go-routine workers
when initialized. One of the worker threads takes a message and processes
it. After unmarshalling and  processing the result is pushed to one of the four channels
for each type of operations. Main queue thread collects from the channels 
and performs the actual operation on the internal queue.

For the operations GetItem and GetAll (opcodes '?' and '*') the server forms a response and 
sends the response back to the client via client's SQS queue.

Server writes output messages to stdout as well as to ./server.log


### Network format (from clients to the server)

SQS message's body is a string, so the actual application's payload is marshalled to a json string.
Depending on the particular request, the message may have fields:

```go
type NetPayload struct {
	OpCode      rune      `json:"op"`
	Key         string    `json:"key"`
	Value       *QPayload `json:"value"`
	ClientQueue string    `json:"ret_queue"`
}
``` 
Notes:
* The _OpCode_ field is mandatory for any type of message.
* _Key_ string keeps item key. It is mandatory for opcodes '+', '-' and '?'.
* _Value_ is marshalled item's value when adding an item (opcode '+'). In the current implementation it is a string.
* _ClientQueue_ defines name of the client's SQS queue where it receives responses.



### Client implementation

The client is expecting a command from the user. Each command should start
with an opcode. For opcodes other than '*' (_GetAllItems_) the command is 
followed by a key or a key/value pair separated by semicolon).
 
Examples:

_AddItem_:
```
+key1:value1
+key2:value2
```

_DelItem_:
```
-key1
-key2
```

_GetItem_:
```
?key1
?key2
```

_GetAllItems_:

```
*
```


### Server command line flags

```
"profile"  --   Sets aws profile name (default: "default" 
"region"   --   Sets AWS region  (default: "eu-west-1")
"qname",   --   Server's SQS queue name ("bl-test-queue")
"threads"  --   Number of threads processing payload (default: 3)
```

Example:
```
./server -profile=default -region=us-east-1 -threads=8
```

### Clients command line flags 

```
"profile"  --   Sets aws profile name (default: "default" 
"region"   --   Sets AWS region  (default: "eu-west-1")
"qname"   --   Server's SQS queue name ("bl-test-queue")
```

Example:
```
./client -profile=default -region=us-east-1 -qname=myserversqs
```


### Compiling and running

from the repository root:

```
go build -o client client/client.go 
go build -o server server/server.go
```

to run:

```
Server:
server/server

Client:
client/client
```


