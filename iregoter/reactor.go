package main

import (
    "fmt"
    "log"
    "net"
)

type eventHandler interface {
    handleEvent(net.Conn, chan<- string)
}

type reactor struct {
    handlers map[string]eventHandler
    events   chan string
}

func newReactor() *reactor {
    return &reactor{
        handlers: make(map[string]eventHandler),
        events:   make(chan string),
    }
}

func (r *reactor) registerHandler(eventName string, handler eventHandler) {
    r.handlers[eventName] = handler
}

func (r *reactor) removeHandler(eventName string) {
    delete(r.handlers, eventName)
}

func (r *reactor) handleEvents() {
    listener, err := net.Listen("tcp", "localhost:8080")
    if err != nil {
        log.Fatal(err)
    }
    defer listener.Close()

    for {
        conn, err := listener.Accept()
        if err != nil {
            log.Println(err)
            continue
        }

        go func() {
            defer conn.Close()

            // Read incoming data from the connection
            buffer := make([]byte, 1024)
            n, err := conn.Read(buffer)
            if err != nil {
                log.Println(err)
                return
            }

            // Send the event to the reactor's event channel
            r.events <- string(buffer[:n])
        }()
    }
}

func (r *reactor) start() {
    for {
        select {
        case eventName := <-r.events:
            handler, ok := r.handlers[eventName]
            if !ok {
                log.Printf("Unknown event: %s\n", eventName)
                continue
            }

            // Handle the incoming event using the event handler
            conn, err := net.Dial("tcp", "localhost:8080")
            if err != nil {
                log.Println(err)
                continue
            }
            defer conn.Close()

            handler.handleEvent(conn, r.events)
        }
    }
}

type echoHandler struct{}

func (h *echoHandler) handleEvent(conn net.Conn, events chan<- string) {
    buffer := make([]byte, 1024)
    n, err := conn.Read(buffer)
    if err != nil {
        log.Println(err)
        return
    }

    message := string(buffer[:n])
    log.Printf("Received message: %s\n", message)

    // Echo the message back to the client
    _, err = conn.Write([]byte(message))
    if err != nil {
        log.Println(err)
        return
    }

    // Send an "echo" event back to the reactor
    events <- "echo"
}

func main() {
    // Create a new reactor
    reactor := newReactor()

    // Register an event handler
    reactor.registerHandler("echo", &echoHandler{})

    // Start handling events
    go reactor.handleEvents()
    reactor.start()
}

