package main

import (
	"bufio"
	"flag"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"
)

var readKeyFile = ioutil.ReadFile
var bashColors = [6]int{31, 32, 33, 34, 35, 36}

type Params struct {
	Ips     []string
	Key     []byte
	User    string
	Command string
	Timeout int
	flag    *flag.FlagSet
	showIp  bool
}

type ErrorParam string

type SshConn struct {
	index     int
	ch        chan string
	ip        string
	command   *string
	sshConfig *ssh.ClientConfig
	showIp    bool
}

func (e ErrorParam) Error() string {
	return fmt.Sprintf("Argument error: %s", string(e))
}

func (p *Params) Parse() (*Params, error) {
	flag := flag.NewFlagSet("rcmd_args", flag.ExitOnError)

	keyPathPtr := flag.String("k", "", "path to PEM key")
	userPtr := flag.String("u", "", "username")
	ipsPtr := flag.String("H", "", "List of host IPs (comma separated)")
	commandPtr := flag.String("c", "", "Command to run")
	timeoutPtr := flag.Int("t", 30, "Timeout in seconds")
	showIpPtr := flag.Bool("show-ip", true, "show host IP for every line")

	flag.Parse(os.Args[1:])
	p.flag = flag

	if *userPtr == "" {
		return nil, ErrorParam("user")
	}

	if *commandPtr == "" {
		return nil, ErrorParam("command")
	}

	key, err := readKeyFile(*keyPathPtr)

	if err != nil {
		return nil, ErrorParam("key path")
	}

	p.Key = key
	p.User = *userPtr
	p.Ips = strings.Split(*ipsPtr, ",")
	p.Command = *commandPtr
	p.Timeout = *timeoutPtr
	p.showIp = *showIpPtr

	return p, nil
}

func (c SshConn) Connect() {
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", c.ip), c.sshConfig)
	if err != nil {
		log.Printf("unable to connect: %v", err)
		return
	}

	session, err := client.NewSession()
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		return
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		log.Printf("unable to acquire stdout pipe: %s", err)
		return
	}

	err = session.Start(*c.command)
	if err != nil {
		log.Printf("unable to execute remote command: %s", err)
		return
	}

	reader := bufio.NewReader(stdout)

	for {
		line, err := reader.ReadString('\n')
		if err != nil || io.EOF == err {
			if err != io.EOF {
				log.Printf("eof: %s", err)
			}
			session.Close()
			break
		}
		if c.showIp {
			c.ch <- fmt.Sprintf("%s%s", c.FormatedIp(), line)
		} else {
			c.ch <- line
		}
	}
}

func (c SshConn) FormatedIp() string {
	n := 16 - len(c.ip)
	color := bashColors[c.index%len(bashColors)]
	padding := strings.Repeat(" ", n)

	return fmt.Sprintf("\033[%dm%s:\033[39m%s", color, c.ip, padding)
}

func sshConfig(p *Params) *ssh.ClientConfig {
	key, err := ssh.ParsePrivateKey(p.Key)
	if err != nil {
		log.Fatalf("Can't parse private key %v", err)
	}

	return &ssh.ClientConfig{
		User:            p.User,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(key)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Second * time.Duration(p.Timeout),
	}
}

func main() {
	params := Params{}
	_, err := params.Parse()
	if err != nil {
		fmt.Println(err)
		params.flag.PrintDefaults()
		os.Exit(1)
	}

	configPtr := sshConfig(&params)

	c := make(chan string)
	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt)

	for i, ip := range params.Ips {
		sshConn := SshConn{
			index:     i,
			ip:        ip,
			ch:        c,
			showIp:    params.showIp,
			command:   &params.Command,
			sshConfig: configPtr,
		}
		go sshConn.Connect()
	}

	for {
		select {
		case msg := <-c:
			fmt.Print(msg)
		case <-s:
			fmt.Println("\nBye!")
			close(c)
			os.Exit(1)
		}
	}
}
