# WOW-using-PoW

World of Wisdom TCP Server:

- Protected from DDOS attacks with the Proof of Work (https://en.wikipedia.org/wiki/Proof_of_work) algorithm.
- Challenge-response protocol is used.
- Chose Hashcash algorithm, because:
    - It is enough for the formulated task
    - Easy to implement
- After Proof Of Work verification, server is sending one of the quotes from https://en.wikisource.org/wiki/The_Doctrine_and_Covenants/Section_89. 
- Docker file is provided both for the server and for the client that solves the POW challenge


# How to trigger

To start both client and server I used docker compose file with specific dependencies between containers:
1. Redis
2. Server
3. Client

```
sudo docker compose up --build
```

After start you will see something like this:
```
... more redis setup
redis   | 1:M 22 Jan 2024 19:21:32.331 * Ready to accept connections tcp
server  | 2024/01/22 19:21:32 starting server...
client  | 2024/01/22 19:21:33 starting client...
client  | 2024/01/22 19:21:33 connected to server:50051
client  | 2024/01/22 19:21:33 running client...
server  | 2024/01/22 19:21:33 handling connection: 172.21.0.3:49732
server  | 2024/01/22 19:21:33 client 172.21.0.3:49732 requests challenge
client  | 2024/01/22 19:21:33 solving puzzle: &{1 3 2024-01-22 19:21:33.118677149 +0000 UTC 172.21.0.3:49732  MTU1Mzk= 0}
client  | 2024/01/22 19:21:35 puzzle solved: &{1 3 2024-01-22 19:21:33.118677149 +0000 UTC 172.21.0.3:49732  MTU1Mzk= 1925454}
client  | 2024/01/22 19:21:35 challenge solution sent to server
server  | 2024/01/22 19:21:35 client solved challenge and requests resource. client: 172.21.0.3:49732, payload {"Version":1,"ZerosCount":3,"Date":"2024-01-22T19:21:33.118677149Z","Resource":"172.21.0.3:49732","Extension":"","Rand":"MTU1Mzk=","Counter":1925454}
server  | 2024/01/22 19:21:35 Solution verified. Sending a word of wisdom.
client  | 2024/01/22 19:21:35 Received quote: Yea, flesh also of beasts and of the fowls of the air, I, the Lord, have ordained for the use of man with thanksgiving; nevertheless they are to be used sparingly;
```