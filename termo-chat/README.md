# termo-chat

termo-chat is a simple chat in a terminal.

![termo-chat](https://github.com/austinov/go-recipes/blob/assets/termo-chat/screenshot.png)

First run the server (default on 8822 port):

```
$ cd termo-chat/server
$ go run main.go
```

Next run the client to book room:

```
$ cd termo-chat/client:
$ go run main.go
```

To join the room run the client in another terminal with room id from first client:

```
$ cd termo-chat/client:
$ go run main.go -room=[room_id]
```

To set port of the server and client used flag "-addr":

```
go run main.go -addr=:8888
```

or

```
$ go run main.go -addr=127.0.0.1:9000
```
