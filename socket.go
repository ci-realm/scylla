package main

import (
	"time"

	macaron "gopkg.in/macaron.v1"
)

type Message struct {
}

func getSocket(ctx *macaron.Context, receiver <-chan *Message, sender chan<- *Message, done <-chan bool, disconnect chan<- int, errorChannel <-chan error) {
	ticker := time.After(30 * time.Minute)
	for {
		select {
		case msg := <-receiver:
			// here we simply echo the received message to the sender for demonstration purposes
			// In your app, collect the senders of different clients and do something useful with them
			sender <- msg
		case <-ticker:
			// This will close the connection after 30 minutes no matter what
			// To demonstrate use of the disconnect channel
			// You can use close codes according to RFC 6455
			// https://tools.ietf.org/html/rfc6455#section-7.4.1
			disconnect <- 1000
		case <-done:
			return
		case err := <-errorChannel:
			logger.Println(err)
		}
	}
}
