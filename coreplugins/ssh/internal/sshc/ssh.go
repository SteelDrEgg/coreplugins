package sshc

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

type Host struct {
	User         string
	Host         string
	Port         string
	Hostname     string
	IdentityFile string
	Timeout      time.Duration
}

type Identity struct {
	KeyPath    string
	Passphrase string
}

// LoadKey loads a private key for SSH authentication.
func LoadKey(key *Identity) (ssh.Signer, error) {
	if key.KeyPath == "" {
		key.KeyPath = "$HOME/.ssh/id_rsa"
	}
	keyPath := os.ExpandEnv(key.KeyPath)

	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("private key file does not exist: %s", keyPath)
	}

	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("read private key from %s: %w", keyPath, err)
	}

	if key.Passphrase != "" {
		signer, err := ssh.ParsePrivateKeyWithPassphrase(keyBytes, []byte(key.Passphrase))
		if err != nil {
			return nil, fmt.Errorf("parse private key with passphrase: %w", err)
		}
		return signer, nil
	}

	signer, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}
	return signer, nil
}

// LoadAuth creates SSH authentication methods from password and identities.
func LoadAuth(password string, identities []*Identity) ([]ssh.AuthMethod, error) {
	var authMethods []ssh.AuthMethod
	if password != "" {
		authMethods = append(authMethods, ssh.Password(password))
	}

	for _, id := range identities {
		if id == nil {
			continue
		}
		signer, err := LoadKey(id)
		if err != nil {
			log.Printf("failed to load key from %s: %v", id.KeyPath, err)
			continue
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no valid authentication methods available")
	}
	return authMethods, nil
}

// Connect creates an SSH connection.
func Connect(host *Host, auth []ssh.AuthMethod) (*ssh.Client, error) {
	sshConfig := &ssh.ClientConfig{
		User:            host.User,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         host.Timeout,
	}
	addr := net.JoinHostPort(host.Hostname, host.Port)

	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", addr, err)
	}
	return client, nil
}
