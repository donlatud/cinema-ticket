package mq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewClient(url string) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("open channel: %w", err)
	}

	client := &Client{conn: conn, ch: ch}
	if err := client.declareQueues(); err != nil {
		_ = client.Close()
		return nil, err
	}

	return client, nil
}

func (c *Client) Channel() *amqp.Channel {
	return c.ch
}

func (c *Client) Close() error {
	if c.ch != nil {
		_ = c.ch.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) declareQueues() error {
	queues := []string{QueueBookingSuccess, QueueBookingTimeout, QueueSeatReleased}
	for _, name := range queues {
		_, err := c.ch.QueueDeclare(name, true, false, false, false, nil)
		if err != nil {
			return fmt.Errorf("declare queue %s: %w", name, err)
		}
	}
	return nil
}
