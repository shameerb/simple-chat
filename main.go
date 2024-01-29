package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

// server

type server struct {
	rooms    map[string]*room
	commands chan command
}

func newServer() *server {
	return &server{
		rooms:    make(map[string]*room),
		commands: make(chan command),
	}
}

func (s *server) newClient(conn net.Conn) *client {
	return &client{
		name:     "anonymous",
		conn:     conn,
		commands: s.commands,
	}
}

func (s *server) run() {
	for cmd := range s.commands {
		switch cmd.id {
		case CMD_NAME:
			s.name(cmd.client, cmd.args)
		case CMD_JOIN:
			s.join(cmd.client, cmd.args)
		case CMD_ROOMS:
			s.listRooms(cmd.client)
		case CMD_MSG:
			s.msg(cmd.client, cmd.args)
		case CMD_QUIT:
			s.quit(cmd.client)
			// default:
		}
	}
}

func (s *server) name(c *client, args []string) {
	if len(args) < 2 {
		c.err(fmt.Errorf("<name> is required as a parameter for /name command"))
		return
	}
	c.name = args[1]
	c.msg(fmt.Sprintf("Hello %s", c.name))
}

func (s *server) join(c *client, args []string) {
	if len(args) < 2 {
		c.err(fmt.Errorf("<room name> is required as a parameter for /join command"))
		return
	}
	name := args[1]
	r, ok := s.rooms[name]
	if !ok {
		r = &room{
			name:    name,
			members: make(map[net.Addr]*client),
		}
		s.rooms[name] = r
	}
	s.quitRoom(c)
	c.room = r
	r.members[c.conn.RemoteAddr()] = c
	r.broadcast(c, fmt.Sprintf("%s joined the room", c.name))
}

func (s *server) listRooms(c *client) {
	var res []string
	for n := range s.rooms {
		res = append(res, n)
	}
	c.msg(fmt.Sprintf("rooms: %s", strings.Join(res, ", ")))
}

func (s *server) msg(c *client, args []string) {
	if len(args) < 2 {
		c.err(fmt.Errorf("<message> required as parameter in /msg command"))
		return
	}
	msg := strings.Join(args[1:], " ")
	c.room.broadcast(c, fmt.Sprintf("%s: %s", c.name, msg))
}

func (s *server) quit(c *client) {
	if c.room != nil {
		r := c.room
		s.quitRoom(c)
		c.msg(fmt.Sprintf("you left the room: %s", r.name))
	}
	c.msg("closing connection")
	c.conn.Close()
}

func (s *server) quitRoom(c *client) {
	if c.room != nil {
		delete(s.rooms[c.room.name].members, c.conn.RemoteAddr())
		s.rooms[c.room.name].broadcast(c, fmt.Sprintf("%s left", c.name))
		c.room = nil
	}
}

// Room

type room struct {
	name    string
	members map[net.Addr]*client
}

func (r *room) broadcast(sender *client, msg string) {
	for _, c := range r.members {
		if c == sender {
			continue
		}
		c.msg(msg)
	}
}

// Client
type client struct {
	name     string
	conn     net.Conn
	room     *room
	commands chan<- command
}

// todo: stop reading from the conn after quit.
func (c *client) readInput() {
	for {
		msg, err := bufio.NewReader(c.conn).ReadString('\n')
		if err != nil {
			log.Printf("failed to read message from %s: %s", c.conn.RemoteAddr(), err)
			return
		}

		msg = strings.Trim(msg, "\r\n")
		args := strings.Split(msg, " ")
		cmd := strings.TrimSpace(args[0])
		switch cmd {
		case "/name":
			c.commands <- command{
				id:     CMD_NAME,
				args:   args,
				client: c,
			}
		case "/join":
			c.commands <- command{
				id:     CMD_JOIN,
				args:   args,
				client: c,
			}
		case "/msg":
			c.commands <- command{
				id:     CMD_MSG,
				args:   args,
				client: c,
			}
		case "/rooms":
			c.commands <- command{
				id:     CMD_ROOMS,
				client: c,
			}
		case "/quit":
			c.commands <- command{
				id:     CMD_QUIT,
				client: c,
			}
		default:
			c.err(fmt.Errorf("unknown command: %s", cmd))
		}
	}
}

func (c *client) err(err error) {
	c.conn.Write([]byte("err: " + err.Error() + "\n"))
}

func (c *client) msg(msg string) {
	c.conn.Write([]byte("> " + msg + "\n"))
}

// Command

type commandID int

const (
	CMD_NAME commandID = iota
	CMD_JOIN
	CMD_ROOMS
	CMD_MSG
	CMD_QUIT
)

type command struct {
	id     commandID
	client *client
	args   []string
}

func main() {
	// create an instance of server and // listen to commands on a go routine in a channel
	s := newServer()
	go s.run()

	// create a tcp server
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("TCP server failed to start %s\n", err)
	}
	defer listener.Close()

	// accept new connections on the tcp port and create a new client
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connections: %s\n", err)
		}

		c := s.newClient(conn)
		// keep a continous loop on a goroutine listening on this client and send the commands to the main chan
		// todo: stop reading from the conn after quit. Use a context to exit the goroutine on a cancel on the context. Next Steps
		go c.readInput()
	}
}
