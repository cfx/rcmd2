## rcmd2

Run commands on remote hosts via ssh.

```
$ go test
$ go build
```

#### Usage:

```
$ rcmd --help
```

#### Example:

```
$ rcmd2 -H ip1,ip2,ip3 -u username -k ~/.ssh/path_to_key -c 'hostname'
```
