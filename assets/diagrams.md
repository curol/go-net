# Diagrams

## Stream

bytes -> NewBuffer(bytes) -> NewReader(buffer) -> HandleRead(reader)

## Connection
// Message -> Connection -> Read -> Parse -> Request -> Route -> Handler -> Response -> Write -> Connection -> Message

## A
```mermaid
flowchart LR

Client[Client]--> |Message| B(Server)


B --> |Read|C{Router}

C -->|GET /path| D[Result 1]

C -->|Two| E[Result 2]

```




## B
```mermaid
flowchart LR

    Client--> |Request Message| Server(Server)

    subgraph Server
        Connection--> |Read Client Message| Router{Router}-->|Route /path|Handler
    end
        Server-->|Response Message|Client
```

## C

```mermaid
flowchart LR
    Client

    subgraph Server
        Listen
        -->|New Connection to client|Connection
        -->|Connection Reader / Writer|Router
        -->Handler
        -->Database
    end

    Client-->|Request Message|Server

    Server-->|Response Message|Client

```

## D

```mermaid
flowchart LR

a --> b & c--> d

id1[(Database)]

id2((This is the text in the circle))

A[Client] -->|Message| B(Server)
```


## Server-Client Sequence
```mermaid
sequenceDiagram
    Server->>Client: Request Message
    Client-->>Server: Response Message
```

