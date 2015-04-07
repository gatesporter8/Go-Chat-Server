//Richard Gates Porter rgp633
//NU Chitter Project 0
//4-5-15

package main

import (
    
    "fmt"
    "net"
    "os"
    "bufio"
    "io"
    "strings"
    "strconv"
)

type Client struct {//organizes all data about a Client into a single structure
    conn net.Conn //connection
    ID string //user id
    ch chan Format //clients channel, HandleMessages takes messages from "messages" and puts them in each Clients channel
}

type Format struct {//organizes message formatting details into structure. helps support private messaging.
    FROM string
    MESSAGE string
    TO string
}

var idAssignmentChan = make(chan string)//creates a channel from which each new client gets a number id (0-"inf", first come first serve)

func (c Client) ReadIntoChan(ch chan <- Format){//allows specific client to input message (written into messages channel)
    b := bufio.NewReader(c.conn)
    for {
        line, err := b.ReadBytes('\n')
        teststr := string(line)
        s := strings.TrimSpace(teststr)
        if err != nil {
            break
        }
        if strings.ContainsAny(s, ":"){
            arr := strings.SplitAfterN(s, ":",2)
            test := arr[0]
            stuff := arr[1]
                        final := strings.Replace(test, " ", "", -1)
            IDGET := strings.Replace(final, ":","",1)
            content := strings.TrimSpace(stuff)
            if final == "all:"{
                format := Format{
                    FROM: c.ID,
                    MESSAGE: content,
                    TO: "blah",
                }
                ch <- format
            }else if final == "whoami:"{
                c.conn.Write([]byte("chitter: "+c.ID+"\n"))
            }else {
                format := Format{
                    FROM: c.ID,
                    MESSAGE: content,
                    TO: IDGET,
                }
                ch <- format
            }
        } else {
            
                format := Format{
                    FROM: c.ID,
                    MESSAGE: s,
                    TO: "blah",
                }
                ch <- format
        }
    }
}

func (c Client) WriteFromChan(ch <- chan Format){//outputs messages in clients channel into clients connection 
    for msg := range ch{
        output := fmt.Sprintf("%s: %s\n", msg.FROM, msg.MESSAGE)
        if msg.TO != "blah" {//private message
            if msg.TO == c.ID {
                _, err := io.WriteString(c.conn, output)
                if err != nil{
                    return
                }
            }
        }else {//everyone should see this
            _, err := io.WriteString(c.conn, output)
            if err != nil{
                return
            }
        }
    }
}    

func HandleConnection(conn net.Conn, messages chan<-Format, incoming chan<-Client, leaving chan<-Client){
    defer conn.Close()
    client_id := <-idAssignmentChan//sets client id queing from channel of ids
    client := Client{
        conn: conn,
        ID: client_id,
        ch: make(chan Format),
    }
    
    incoming <- client//adds client to channel
    
    defer func(){

        leaving <-client
    }()

    go client.ReadIntoChan(messages)//allows client to write to other clients
    client.WriteFromChan(client.ch)//outputs messages from other clients. 
}

func IdManager() {//id generator puts ids in idAssignmentChan
    var i uint64
    for i = 0; ; i++{
        idAssignmentChan <- strconv.FormatUint(i,10)
    }
}


func handleMessages(messages <-chan Format, incoming <- chan Client, leaving <- chan Client){//fowards messages from message channel into client channel, and updates clients map.
    
    clients := make(map[net.Conn]chan<-Format)

    for {
        select {
        case msg := <-messages:
            for _, ch := range clients {
                go func(mch chan <- Format) {mch <- msg}(ch)
            }
        case client := <-incoming:
            clients[client.conn] = client.ch
        case client := <-leaving:
            delete(clients, client.conn)
        }
    }
}

func main() {
    
    if len(os.Args) < 2 {//check to make sure the proper amount of inputs
        fmt.Fprintf(os.Stderr, "Usage: chitter <port-number>\n")
        os.Exit(1)
        return
    }

    portNum := os.Args[1] //get port number from cmdline input

    server, err := net.Listen("tcp", ":"+portNum)//creates server
    if err != nil {
        fmt.Fprintln(os.Stderr, "Can't connect to port")
        os.Exit(1)
    }
    
    go IdManager()

    messages := make(chan Format)
    incoming := make(chan Client)//facilitates addition into clients map
    leaving := make(chan Client)//facilitates deletion from clients map
    
    go handleMessages(messages, incoming, leaving)//starts handmessages goroutine

    fmt.Println("Listening on port", os.Args[1])

    for {
        conn, _ := server.Accept()
            go HandleConnection(conn, messages, incoming, leaving)//starts hconnection goroutine
    }
    
}