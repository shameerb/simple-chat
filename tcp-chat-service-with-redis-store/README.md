## TCP Chat service
TCP chat service using Redis Pub-Sub, gRPC, Go.

### Components
- server
    - starts a grpc server to accept messages from the clients
    - grpc methods
        - connect, disconnect, message, list users
    - connects to the redis server for managing users and storing messages.
    
- client 
    - makes a client connection (connect request) to the grpc server (server)
    - listens to messages on the redis channel
    - waits for message on the command prompt to be sent to the server
    - disconnect request to server on exit

- redis 
    - stores the messages from each of the client which needs to be broadcasted to all subscribed clients.

### Feature Enhancements
- when a user exists, call the disconnect rpc call.
    - server should have an active connection (heartbeat system) to check for idle users.
    - Also, the server can end the connection as well if required.
- interface to change the storage layer from redis to another store (file, sql etc)
- Allow for multiple rooms
- Allow for users to be part of multiple rooms
- A message will be sent to a specific user and room
- Scale the number of rooms, users and messages
- The messages will be displayed when the user comes online (connects)