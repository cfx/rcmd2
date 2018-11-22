package main

import (
	"flag"
	"fmt"
	"strings"
)

type Params struct {
	Ips     []string
	KeyPath string
	User    string
}

type ErrMissingParam string

type IPaddr int32

func (e ErrMissingParam) Error() string {
	return fmt.Sprintf("Missing required argument: %s", string(e))
}

func (p *Params) Parse() (*Params, error) {
	keyPathPtr := flag.String("k", "", "path to PEM key")
	userPtr := flag.String("u", "", "username")
	ipsPtr := flag.String("H", "", "List of host IPs")

	flag.Parse()

	if *userPtr == "" {
		return p, ErrMissingParam("User")
	}

	if *keyPathPtr == "" {
		return p, ErrMissingParam("Key Path")
	}

	p.KeyPath = *keyPathPtr
	p.User = *userPtr
	p.Ips = strings.Split(*ipsPtr, ",")

	return p, nil
}

func main() {
	params := &Params{}
	_, err := params.Parse()

	if err == nil {
		for _, v := range params.Ips {
			fmt.Println(v)
		}

	} else {
		fmt.Println(err)
	}
}
