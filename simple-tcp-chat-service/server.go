package main

import (
	"fmt"
	"net"
	"strings"
)

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
