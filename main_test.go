package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

type MockReadKeyFile struct{ Str string }
type ErrorMockReadKeyFile string

func (e ErrorMockReadKeyFile) Error() string { return string(e) }

func (k MockReadKeyFile) ReadFile(path string) ([]byte, error) {
	if path == "ok" {
		buf := bytes.NewBufferString(k.Str)
		return ioutil.ReadAll(buf)
	} else {
		return nil, ErrorMockReadKeyFile(path)
	}
}

func TestParams(t *testing.T) {
	mock := MockReadKeyFile{"foo"}
	readKeyFile = mock.ReadFile

	errorCases := []struct {
		Name, Ips, Key, User, Command, Expected string
	}{
		{
			"missing user param",
			"-H=8.8.8.8",
			"-k=ok",
			"",
			"-c=hostname",
			"Argument error: user",
		},
		{
			"missing key param",
			"-H=8.8.8.8",
			"-k=err",
			"-u=user",
			"-c=hostname",
			"Argument error: key path",
		},
		{
			"missing command param",
			"-H=8.8.8.8",
			"-k=ok",
			"-u=user",
			"",
			"Argument error: command",
		},
	}

	for _, tc := range errorCases {
		t.Run(tc.Name, func(t *testing.T) {
			os.Args = []string{
				"cmd", tc.Ips, tc.Key, tc.User, tc.Command,
			}

			params := Params{}
			_, err := params.Parse()
			if err.Error() != tc.Expected {
				t.Fatal("expected:", tc.Expected, "got:", err)
			}
		})
	}

	t.Run("all arguments valid", func(t *testing.T) {
		expected := Params{
			Ips:     []string{"8.8.8.8"},
			Key:     []byte("foo"),
			User:    "user",
			Command: "hostname",
			Timeout: 30,
			showIp:  true,
		}

		os.Args = []string{
			"cmd",
			"-H=8.8.8.8",
			"-k=ok",
			"-u=user",
			"-c=hostname",
		}

		params := Params{}
		params.Parse()

		if params.Ips[0] != expected.Ips[0] {
			t.Fatal("expected:", expected.Ips, "got:", params.Ips)
		}

		if string(params.Key) != string(expected.Key) {
			t.Fatal("expected:", expected.Key, "got:", params.Key)
		}

		if params.User != expected.User {
			t.Fatal("expected:", expected.User, "got:", params.User)
		}

		if params.Command != expected.Command {
			t.Fatal("expected:", expected.Command,
				"got:", params.Command)
		}

		if params.Timeout != expected.Timeout {
			t.Fatal("expected:", expected.Timeout,
				"got:", params.Timeout)
		}

		if params.showIp != expected.showIp {
			t.Fatal("expected:", expected.Timeout,
				"got:", params.Timeout)
		}

	})
}
