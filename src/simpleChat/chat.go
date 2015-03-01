package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

var _ = fmt.Println

var Log *log.Logger

type Client struct {
	outgoing chan string
	reader   *bufio.Reader
	writer   *bufio.Writer
	name     string
	ip       string
}

func dealBr(s string) string {
	return strings.TrimRight(s, "\n")
}

func (client *Client) Read() {
	for {
		line, err := client.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return
			}
			Log.Println(err)
			delete(onlineClients, client.name)
			brodCast(fmt.Sprintf("user %s leave\n", client.name))
			return
		}
		if len(line) > 2 && line[0] == '/' {
			msg := line[1:]
			sp := strings.SplitN(msg, " ", 2)
			Log.Println(sp)
			if len(sp) == 2 {
				name := sp[0]
				msg := sp[1]
				msg = fmt.Sprintf("msg from %s to %s : %s\n", client.name, name, msg)
				Log.Println(msg)
				to, ok := onlineClients[name]
				if ok {
					to.outgoing <- msg
				} else {
					msg := "no such user online"
					Log.Println(msg)
					client.outgoing <- msg
				}
			}
		} else {
			msg := fmt.Sprintf("%s: %s\n", client.name, line)
			Log.Println(dealBr(msg))
			brodCast(msg)
		}
	}
}

func (client *Client) Write() {
	for data := range client.outgoing {
		client.writer.WriteString(dealBr(data) + "\n")
		client.writer.Flush()
	}
}

func (client *Client) Listen() {
	go client.Read()
	go client.Write()
}

func brodCast(msg string) {
	Log.Println("brodCast", dealBr(msg))
	for _, v := range onlineClients {
		v.outgoing <- dealBr(msg) + "\n"
	}
}

func NewClient(connection net.Conn) *Client {
	writer := bufio.NewWriter(connection)
	reader := bufio.NewReader(connection)
	writer.WriteString("Input you name:")
	writer.Flush()
	name, _ := reader.ReadString('\n')
	name = strings.Replace(name, "\n", "", -1)
	client := &Client{
		outgoing: make(chan string),
		reader:   reader,
		writer:   writer,
		ip:       connection.RemoteAddr().String(),
		name:     name,
	}
	msg := fmt.Sprintf(
		`You name:%s
You ip:%s
type "/name msg" to chat 
type "msg" to broadcast
Online user %d:
`,
		name, client.ip, len(onlineClients))
	for _, v := range onlineClients {
		line := fmt.Sprintf("  name: %s , ip: %s\n", v.name, v.ip)
		msg += line
	}
	writer.WriteString(msg)
	writer.Flush()
	msg = fmt.Sprintf("New user %s %s\n", name, client.ip)
	Log.Println(dealBr(msg))
	brodCast(msg)
	client.Listen()
	return client
}

var onlineClients map[string](*Client)
var port = flag.Int("port", 6666, "port=6666 ")

func main() {
	onlineClients = make(map[string](*Client), 0)
	listener, err := net.Listen("tcp", "0.0.0.0:"+strconv.Itoa(*port))
	if err != nil {
		Log.Fatalln(err)
	}
	log.Println("simpleChat run on ", "0.0.0.0:"+strconv.Itoa(*port))
	log.Println("please type \"nc serverip youport\" to connect it")
	log.Println("for example: nc localhost 6666")
	for {
		conn, _ := listener.Accept()
		c := NewClient(conn)
		onlineClients[c.name] = c
	}
}
func init() {
	Log = log.New(os.Stdin,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lshortfile)
	flag.Parse()
}
