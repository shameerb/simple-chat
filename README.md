## TCP chat service
- There are 2 version of this.   
    - The first one simply starts a server and you can connect to it directly using netcat command. The clients are stored in memory and messages are broadcast by the server.  
    - The second version has a seperate server and client service. It uses gRPC to communicate and uses redis as a broker service to store the messages and users. The users and messages are stored in redis and uses a pub - sub model to send and fetch messages.  