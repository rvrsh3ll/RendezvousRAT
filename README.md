# Information

This repository contains two minimal proof-of-concept RAT's utilizing GO, based on the examples found at [go-libp2p-examples](https://github.com/libp2p/)

Currently, the RAT is inteded for linux hosts. However, they can be built for windows by modifying the go build variable.
You can set the environment variable in linux to tell GO to compille for windows. "env GOOS=windows" and change the command in the project from "sh -c" to "cmd.exe /c". This will be implemented in a later version.


# RendezvousRAT

 Self-healing proof-of-concept RAT utilizing libp2p. 

 ## Building RendezvousRAT

 This project consists of a client and server. The server ip and port must be accessible to the client. Build the client and server binaries separtely. 

 ### Server

 From the server directory, "go build .". This will build the server binary.

 ### Client

 From the client directory, "go build .". This will build the client binary.

 ## Execution

 Execute the server and display help "./server --help".

 Execute the server with a specific rendezvous string "./server -rendezvous S3cr3tK3y"

 Execute the client with the same rendezvous string "./client -rendezvous S3cr3tK3y"


# pubsub_rat

Simple RAT that utilizes pubsub channels for communication. Works on LAN only.

## Building pubsub_rat

From the pubsub_rat directory, "go build .". This will build the "chat" binary. 

## Execution

Execute the binary on two hosts within the same subnet.



