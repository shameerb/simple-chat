package main

import "net"

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
