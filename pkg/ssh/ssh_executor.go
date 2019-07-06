package ssh

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/theskyinflames/sshexecutor/pkg/config"
	"golang.org/x/crypto/ssh"
)

type (
	SSHExecutor struct {
		host     string
		port     int
		user     string
		password string // It's also used as sudoer password
		client   *ssh.Client

		cfg *config.Config
		log *logrus.Logger
	}
)

var (
	mtx      = &sync.Mutex{}
	sudoFunc = func(sudoerPassword string, in io.Writer, output *bytes.Buffer, endChan chan struct{}, errChan chan<- error) {
		for {
			select {
			case <-endChan:
				break
			default:
				mtx.Lock()
				if output.Len() > 0 {
					msg := string(output.Bytes())
					if strings.Contains(msg, "[sudo] ") {
						_, err := in.Write([]byte(sudoerPassword + "\n"))
						if err != nil {
							if err != io.EOF {
								errChan <- errors.Wrap(err, fmt.Sprintf("some went wrong when trying remote sudo"))
							}
						}
					}
				}
				mtx.Unlock()
			}
		}
	}

	execFunc = func(session *ssh.Session, command string, errChan chan<- error) {
		err := session.Run(command)
		if err != nil {
			errChan <- errors.Wrap(err, fmt.Sprintf("some went wrong when trying execute remote command"))
		}
	}
)

func NewSSHExecutorServer(host string, port int, user, password string, cfg *config.Config, log *logrus.Logger) *SSHExecutor {
	return &SSHExecutor{host: host, port: port, user: user, password: password, cfg: cfg, log: log}
}

func (s *SSHExecutor) getAuthMethod() (ssh.AuthMethod, error) {
	return ssh.Password(s.password), nil
}

func (s *SSHExecutor) Connect(timeout time.Duration) error {
	var (
		err                  error
		connTimeout          = s.cfg.DefaultSSHCommandTimeout
		maxConnectAttempts   = 5
		connCheckingInterval = 500 * time.Millisecond
	)

	if int(timeout.Seconds()) > 0 {
		connTimeout = timeout
	}

	authMethod, err := s.getAuthMethod()
	if err != nil {
		return err
	}
	sshConfig := &ssh.ClientConfig{
		Timeout: connTimeout,
		User:    s.user,
		Auth:    []ssh.AuthMethod{authMethod},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	host := fmt.Sprintf("%s:%d", s.host, s.port)
	attempt := 0
	for {
		attempt++
		s.client, err = ssh.Dial("tcp", host, sshConfig)
		if err == nil {
			return nil
		}

		s.log.Error(err.Error())

		if attempt == maxConnectAttempts {
			break
		}

		s.log.WithField("name", host).Warn("Waiting for SSH service on host...")
		time.Sleep(connCheckingInterval)
	}

	return fmt.Errorf("it has not been possible to connect to %s by ssh on port %s", host)
}

func (s *SSHExecutor) Close() error {
	return s.client.Close()
}

func (s *SSHExecutor) PrepareStreams(session *ssh.Session, rsOut, rsErr io.Writer) (stdout io.Reader, stderr io.Reader, err error) {

	if rsOut != nil {
		stdout, err = session.StdoutPipe()
		if err != nil {
			return nil, nil, fmt.Errorf("Unable to setup stdout for session: %v", err)
		}
		go io.Copy(rsOut, stdout)
	}

	if rsErr != nil {
		stderr, err = session.StderrPipe()
		if err != nil {
			return nil, nil, fmt.Errorf("Unable to setup stderr for session: %v", err)
		}
		go io.Copy(rsErr, stderr)
	}

	return
}

func (s *SSHExecutor) Execute(command string) (string, string, error) {
	session, err := s.client.NewSession()
	if err != nil {
		return "", "", err
	}
	defer session.Close()

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	err = session.RequestPty("xterm", 80, 40, modes)
	if err != nil {
		return "", "", err
	}

	// Capture stdout and stderr from remote server
	var rsOut bytes.Buffer
	var rsErr bytes.Buffer

	stdOut, stdErr, err := s.PrepareStreams(session, &rsOut, &rsErr)
	if err != nil {
		return "", "", err
	}

	// Set stdin to provide sudo passwd if it's necessary
	stdIn, _ := session.StdinPipe()

	errChan := make(chan error, 10) //errors chan
	go func(log *logrus.Logger) {
		for {
			err := <-errChan
			log.WithField("err", err.Error()).Error("some weng wrong wen traying to execute a command by ssh")
		}
	}(s.log)

	endChan := make(chan struct{}) // sudo func ending chan
	go sudoFunc(s.password, stdIn, &rsOut, endChan, errChan)

	// Execute the remote command
	execFunc(session, command, errChan)
	err = waitForReaderEmpty(stdOut, &rsOut)
	if err != nil {
		return "", "", err
	}
	err = waitForReaderEmpty(stdErr, &rsErr)
	if err != nil {
		return "", "", err
	}
	close(endChan)

	return rsOut.String(), rsErr.String(), nil
}

func waitForReaderEmpty(reader io.Reader, buff *bytes.Buffer) error {
	var (
		b = make([]byte, 1000)
	)

	for {
		mtx.Lock()
		n, err := reader.Read(b)
		mtx.Unlock()
		if n == 0 || err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		buff.Write(b[:n])
	}
	return nil
}
