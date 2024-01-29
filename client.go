package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

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
