## TCP Chat service
Learn to build a TCP chat service using Redis Pub-Sub, gRPC, Go.

### Enhancements
- Allow for multiple rooms
- Allow for users to be part of multiple rooms
- Scale the number of rooms, users and messages.

### Notes
- How does Redis help? 
    - The data is persisted across service restarts
    - The server doesnt need to do a push across all users (hard to scale). 
        - the users/clients can pull from the redis when they are online.


### Changes required
- when a user exists, call the disconnect rpc call.
    - server should have an active connection (heartbeat system) to check for idle users.
    - Also, the server can end the connection as well if required.
- interface to change the storage layer. 
