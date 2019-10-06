# Realtime N-Player Blackjack
An implementation of Blackjack by [Morgan Gallant](https://morgangallant.com) for his [KP Fellows Application](https://fellows.kleinerperkins.com/).

As part of the application process, I was asked to complete a small optional engineering challenge to build a CLI-based card game, like Blackjack or Solitare. **So I did just that, with a twist**. Rather than single player implementation, my implementation works in realtime for any number of players which may or may not be on different networks. This was accomplished using HTTP2 streaming via gRPC. The rest of this document describes the inner workings of both the server and client implementations, and some of the design choices behind each.

I think it's important to note that this same implementation is highly transferrable to a real world scenario, since any realtime application such as Slack or Facebook Messenger, would use an implementation very similar to this. I personally love this kind of stuff, so it was very fun extending the KP engineering challenge to align more with my experience and interests.

### Installation / Usage

This application has two components, the `server` and the `client`. In order for the application to work correctly, the server must be running in addition to any number of clients. Your best bet is to open multiple terminals, one for the server and another for a client. You can open multiple client terminals if you want!

To run, you will need to build the Go code. To install Go, follow the instructions [here](https://golang.org/doc/install). Once you have this downloaded, you can do the following:
```
git clone https://github.com/MorganGallant/realtime-n-player-blackjack
cd realtime-n-player-blackjack

# In one terminal
cd server
go run .

# In another terminal(s)
cd client
go run
```
If there are any issues running the program, please send me an [email](mailto:morgan@morgangallant.com) and I can help out. No protocol buffer compilation is required, since I've included the generated Go code in the repository. If you wish to re-generate these files, execute `./scripts/generate_pb.sh`.

### Deep Dive: Client Implementation

This application was designed to ensure that the client implementation is as simple as possible, meaning that most of the application logic is done server side. The client does the following:
- Gets the username from the player.
- Subscribes to messages from the server.
  - If the server tells the client to print a message, the client prints it.
  - If the server tells the client to get a user action, the client asks the user for an action and sends it to the server.

Funny enough, that's kind of it. Upon initialization, the client makes a connection with the server (hosted on `localhost:9212`) and then asks the user for their username. The client then makes a subscription request to the server, which may or may not reject the request based on the uniqueness of the username chosen. Once subscribed, the client simply waits for server messages and executes the instructions sent by the server.

### Deep Dive: Server Implementation

The server does the brunt of the work with regards to running the game of blackjack, and managing any number of subscribed players. Golang channels were used extensively for reliable/safe cross-goroutine communication. For those unfamiliar with Go, a goroutine can be thought of as a lightweight thread.

##### Player Management

The server maintains an array of players, which is used to store player/game data. Whenever an incoming subscription request is recieved, we first check if the username is unique (if not, we return an error). If the username is OK, we add a new player to the array of players. Each player structure contains a `MessageQueue` field, which is a buffered channel of `rpc.ServerResponse` objects. The goroutine which initally responded to the users subscription request is used to relay incominng messages from this channel to the client. This allows any other goroutine in the server application to send a message to any currently subscribed player by sending a message through the player's message queue.

##### Incoming Player Actions

An important feature in the game of blackjack is to allow the player to choose to either `hit`, `stand` or `leave` the game. Thus, we must be able to notify the client when to send us an action, and be able to recieve that action. When we want the user to send us an action, we simply send a message to the players message queue telling them to get the users input. The client then does this, and proceeds to send the users choice to the `PerformAction` RPC endpoint, along with their username to identify themselves. All incoming messages to this endpoint are fed into the `IncomingMessages` buffered channel, which can be read by the game loop.

##### Blackjack Gameloop

As you can tell, the brunt of the work for this application was done to enable the client/server communication to run smoothly. Therefore, the actual game of blackjack has been simplified in order to stay within the time guidelines of the project. There is no betting, and the cards have been simplified to numbers. Aces aren't 1 or 11, they're 1. Of course, if this was production grade software, a lot more time would be spent on the game of blackjack itself, rather than focusing on the network communication. However, for the scope of this project, I thought it was valid to simplify blackjack in order to showcase more of my interest and experience in client/server systems using RPC.

The blackjack gameloop runs in a seperate goroutine and uses both the publishing messages to subscribers functionality, in addition to the functionality of asking for user input. The loop starts by ensuring that there is atleast one player subscribed to the game, otherwise, it waits. If at any time a player leaves, the game is restarted. The game proceeds as normal, the dealers hand is dealt, as is each players hand. Then, we proceed through the array of players asking each one to give us user input which constitutes their turn. For any given turn, a player may be asked to give multiple choices since they could `hit` more than once. Once all the players have done, the dealer is then forced to `hit` until they have above or equal to 17, or bust. Once this is completed, we calculate who won or lost and restart the game.

### Tooling / Third-Party Libraries Used

This application leverages [Google's GRPC](https://github.com/grpc/grpc-go) framework for RPC (remote procedure calls) between the client and server application. This is accompanied by [Google's Protobuf](https://github.com/protocolbuffers/protobuf) interchange format for data. This allowed me to define a service definition file ([link](rpc.proto)) to describe how the client / server will talk to one another. I use gRPC a ton for personal projects, and I think it is a great library a wide variety of use cases.

gRPC was chosen over a conventional REST type architecture due to the requirement of the server sending messages in realtime back to the client. gRPC provides a great way of doing this via HTTP2 streaming, whereas in the HTTP1.1 REST world, a polling loop would've been required client-side to check for any queued messages. This would be OK, but certainly not as scalable or idiomatic as the gRPC method.


