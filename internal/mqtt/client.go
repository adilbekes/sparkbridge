package mqtt

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"time"

	mqttlib "github.com/eclipse/paho.mqtt.golang"
)

type Config struct {
	BrokerURL          string
	ClientID           string
	Username           string
	Password           string
	KeepAlive          int
	QoS                byte
	InsecureSkipVerify bool
}

type Client struct {
	client mqttlib.Client
	qos    byte
}

func New(cfg Config, onMessage mqttlib.MessageHandler) (*Client, error) {
	if cfg.BrokerURL == "" || cfg.ClientID == "" {
		return nil, errors.New("mqtt broker_url and client_id are required")
	}
	opts := mqttlib.NewClientOptions().AddBroker(cfg.BrokerURL).SetClientID(cfg.ClientID).SetAutoReconnect(true)
	opts.SetKeepAlive(time.Duration(cfg.KeepAlive) * time.Second)
	opts.SetTLSConfig(&tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: cfg.InsecureSkipVerify})
	if cfg.Username != "" {
		opts.SetUsername(cfg.Username)
	}
	if cfg.Password != "" {
		opts.SetPassword(cfg.Password)
	}
	if onMessage != nil {
		opts.SetDefaultPublishHandler(onMessage)
	}
	cli := mqttlib.NewClient(opts)
	if tok := cli.Connect(); tok.Wait() && tok.Error() != nil {
		return nil, fmt.Errorf("connect mqtt: %w", tok.Error())
	}
	return &Client{client: cli, qos: cfg.QoS}, nil
}

func (c *Client) Publish(ctx context.Context, topic string, payload []byte) error {
	tok := c.client.Publish(topic, c.qos, false, payload)
	if !tok.WaitTimeout(10 * time.Second) {
		return context.DeadlineExceeded
	}
	return tok.Error()
}

func (c *Client) Close() { c.client.Disconnect(250) }
