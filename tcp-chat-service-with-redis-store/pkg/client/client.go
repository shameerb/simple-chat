package client

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis"
	"github.com/shameerb/tcp-chat-redis/pkg/common"
	pb "github.com/shameerb/tcp-chat-redis/pkg/grpcapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// connect to redis to read messages, start a grpc connection with server to register and send commands, read from the buffer continuously,
// you will have 3 go routines. 1-> listen to commands from user stdio,  2-> listen to messages from redis, 3 -> select waiting for ctx done, messages or listen chan.

type Client struct {
	redisAddr        string
	redis            *redis.Client
	pubsub           *redis.PubSub
	serverAddr       string
	chatServerConn   *grpc.ClientConn
	chatServerClient pb.ChatServiceClient
	ctx              context.Context
	cancel           context.CancelFunc
	rcvChannel       chan string
	user             string
	writer           io.Writer
	// wg               sync.WaitGroup
}

func NewClient(redisAddr, serverAddr, user string) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		redisAddr:  redisAddr,
		serverAddr: serverAddr,
		ctx:        ctx,
		cancel:     cancel,
		rcvChannel: make(chan string, 1),
		writer:     os.Stdout,
		user:       user,
	}
}

func (c *Client) Init() error {
	var err error
	// initialize redis
	c.redis = redis.NewClient(&redis.Options{
		Addr:     c.redisAddr,
		Password: "",
		DB:       0,
	})
	// Check the redis connection is working.
	_, err = c.redis.Ping().Result()
	if err != nil {
		log.Printf("Error connecting to redis: %s", err.Error())
		return err
	}

	c.pubsub = c.redis.Subscribe(common.CHANNEL)
	log.Printf("listening to redis on %s", c.redisAddr)

	// dial to the server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c.chatServerConn, err = grpc.DialContext(ctx, c.serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Printf("failed to establish connection with server: %s", err)
		return err
	}
	c.chatServerClient = pb.NewChatServiceClient(c.chatServerConn)

	scanner := bufio.NewScanner(os.Stdin)
	// todo: initialize the user. Block until the user is set. Ideally put a timeout for how long you can wait the client.
	if c.user == "" {
		c.initUser(scanner)
	}

	// listen to commands from command line
	go c.listenInputMessage(scanner)

	// wait for messages on the command channel or message from the redis channel. Act accordingly.
	go c.process()
	c.awaitShutdown()
	return nil
}

func (c *Client) listenInputMessage(scanner *bufio.Scanner) {
	for scanner.Scan() {
		fmt.Println("Reading message ...")
		input_msg := scanner.Text()
		c.rcvChannel <- input_msg
	}
}

func (c *Client) process() {
	// c.wg.Add(1)
	// defer c.wg.Done()
	redisChannel := c.pubsub.Channel()
	for {
		select {
		case <-c.ctx.Done():
			return
		case msg := <-redisChannel:
			fmt.Println("got a message on redis channel")
			req := &pb.Message{
				User: c.user,
				Msg:  msg.Payload,
			}
			_, err := c.chatServerClient.Chat(c.ctx, req)
			if err != nil {
				log.Printf("could not send message to chat server: %s", err)
				panic(err)
			}
		case msg := <-c.rcvChannel:
			c.writer.Write([]byte(msg))
		}
	}
}

func (c *Client) initUser(scanner *bufio.Scanner) {

	for {
		c.write("> Enter a username")
		scanner.Scan()
		user := scanner.Text()
		req := &pb.ConnectRequest{
			User: user,
		}
		if _, err := c.chatServerClient.Connect(c.ctx, req); err != nil {
			c.writer.Write([]byte(err.Error()))
			continue
		}
		// successfully set the user
		c.user = user
		break
	}
}

func (c *Client) write(msg string) {
	c.writer.Write([]byte(msg))
}

func (c *Client) awaitShutdown() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	// wait until you get an interrupt signal
	<-stop
	c.stop()
}

func (c *Client) stop() {
	// call cancel for the context
	c.cancel()
	// ideally wait for the goroutines to finish.
	// c.wg.Done()
	c.chatServerConn.Close()
	c.redis.Close()
	log.Println("Stopping client service..")
}
