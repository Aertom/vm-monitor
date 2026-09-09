package ssh

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

type Client struct {
	config *ssh.ClientConfig
	addr   string
}

func NewClient(host, user, privateKeyPath string, port int, timeout time.Duration) (*Client, error) {
	key, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read private key: %w", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %w", err)
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		// TODO: remplacer par un vrai callback avec known_hosts en production
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         timeout,
	}

	return &Client{
		config: config,
		addr:   net.JoinHostPort(host, fmt.Sprintf("%d", port)),
	}, nil
}

func (c *Client) Run(cmd string) (string, error) {
	conn, err := ssh.Dial("tcp", c.addr, c.config)
	if err != nil {
		return "", fmt.Errorf("dial %s: %w", c.addr, err)
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		return "", fmt.Errorf("new session: %w", err)
	}
	defer session.Close()

	var stdout bytes.Buffer
	session.Stdout = &stdout

	if err := session.Run(cmd); err != nil {
		return "", fmt.Errorf("run %q: %w", cmd, err)
	}

	return stdout.String(), nil
}
