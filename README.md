[![Go Report Card](https://goreportcard.com/badge/github.com/theskyinflames/sshexecutor)](https://goreportcard.com/report/github.com/theskyinflames/sshexecutor)

# SSH Executor
This is a service to execute receipts (sequence of shell commands)  to remote servers by ssh. **Included sudo commands !!!** In addition, it returns in its response the standard-output, as well as the standard-error for each executed command. 
Every call to the service can launch so many commands as it's required.

To make it work, **you don't need to install any agents in the target servers**. The only you have to do is create a user account in target servers. In addition, if the recipes you want to execute in these servers include sudo commands, this user account must have sudo privileges. Basically, Sudo privileges are required only if you want to execute a recipe that includes sudo commands. Done that, the service will use this user account to connect by SSH to the remote server, It also will be used to make the remote sudo if the executing command requires it.

## Security aspects
By using SSH to make the recipes remotely, you have all security context SSH provides. To make a remote server able to be operated by this service, it must to have the user account the SSH Executor service uses to connect. In addition, if the recipes to be executed in this server include sudo commands, the user account has to be suoder.

On another hand, future versions of this service will allow the use of key exchange as password method replacement. It's pending to be coded.

Saying that still is necessary to fill the SSH user password as an environment variable for the service. You should inject it into the container when it starts in a safe way.

## Not supported commands
It's not supported the shell commands with 'su', like 'sudo su *another user*'. It's because the 'su' command starts a new shell process for the new user. When that occurs, these new shell has its own stdout,stdin and stderr. Which can't be captured by the ssh client. From there execution flux is lost and the recipe hangs.

## Configuration
There are six environment variables which must be set to make the service run:
* *SSH_EXECUTOR_API_HOST*: Service API host
* *SSH_EXECUTOR_API_PORT*: Service API port
* *SSH_EXECUTOR_API_DEFAULT_SSH_TIMEOUT*: SSH connection timeout in seconds
* *SSH_EXECUTOR_USER*: Login used for connect by ssh to the remote server
* *SSH_EXECUTOR_PASSWORD*: Password used for ssh connection, also to provide the sudo password

## End points
This service published two endpoints:
* *[host:port]/check* to allow service status checking (GET)
* *[host:port]/runreceipt* to execute a recipe (a list of commands) in a remote server by SSH (POST)

## Rq message
The rq message is a JSON message with these fields:
* *host*: Target server
* *port*: SSH port for the target server
* *recipe*: List of commands to be executed in the target server

## Rs message
The rq message is a JSON message with these fields:
* *response*: The list os the stdout content for each executed command of the recipe
* *responseErr*: If some of the commands has written by the stderr, it will in this field
* *error*: If someone has gone wrong when trying to execute the recipe. For exemple, login fails

## Example without sudo
Rq:
```json
{
  "host":"myhost.mydomain.com",
  "port":22,
  "recipe":[
  	  "date",
  	  "uname -a",
  	  "df -h > myfile.txt",
  	  "cat myfile.txt"
    ]
}
```
Rs:
```json
{
    "response": [
        "dv jun 21 03:50:55 CEST 2019\r\n",
        "Linux mx 4.19.0-5-amd64 #1 SMP Debian 4.19.37-2~mx17+1 (2019-05-15) x86_64 GNU/Linux\r\n",
        "",
        "S. fitxers          Mida En ús Lliure  %Ús Muntat a\r\nudev                3,8G     0   3,8G   0% /dev\r\ntmpfs               777M  1,4M   776M   1% /run\r\n/dev/mapper/rootfs  108G  7,3G    95G   8% /\r\ntmpfs               5,0M  4,0K   5,0M   1% /run/lock\r\ntmpfs               1,6G   58M   1,5G   4% /run/shm\r\n/dev/sda1           487M   82M   376M  18% /boot\r\ncgroup               12K     0    12K   0% /sys/fs/cgroup\r\ntmpfs               777M  4,0K   777M   1% /run/user/115\r\ntmpfs               777M   24K   777M   1% /run/user/1000\r\n"
    ],
    "responseErr": "",
    "error": ""
}
```

## Example with sudo recipe
Rq:
```json
{
  "host":"192.168.1.39",
  "port":22,
  "recipe":[
  	  "sudo ls -lart /root/.bashrc",
  	  "sudo date > myfile.txt",
  	  "sudo cat myfile.txt"
    ]
}
```

Rs:
```json
{
    "response": [
        "-rw-r--r-- 1 root root 570 gen 31  2010 /root/.bashrc\r\n",
        "",
        "dv jun 21 04:04:40 CEST 2019\r\n"
    ],
    "responseErr": "",
    "error": ""
}
```

## Running it !
The service is dockerized. So if you start it as a Docker container, only need to do:
```sh
    make docker-build
    docker-compose up
```

Alternatively, you also can run as a local service. In this case, you must do:
```sh
    make build
    cmd
```

## Architecture
This service has been coded using [gin-gonic](https://gin-gonic.com/) to build the REST API, and the [ssh Go package](https://godoc.org/golang.org/x/crypto/ssh) to build the SSH logic.

I've also used [Alpine Linux](https://alpinelinux.org) to build the Docker images, achieving really light ones:
```sh
❯ docker image list
REPOSITORY                TAG                 IMAGE ID            CREATED             SIZE
sshexecutor               latest              de80807f2fc3        18 hours ago        23.8MB
```

### Packages
These are the packages that conforms the service:
* *cmd*: Is the main package
* *config*: Service configuration taken from environment variables
* *http*: REST api. It incudes router and controller
* *model*: Domain models. In this case, the recipe to be executed
* *service*: The application service. It takes care on execute the recipe using the SSH adapter, take the response and return it to the controller
* *shared*: This package provides interfaces which are shared between several packages. On this way I avoid to duplicate interfaces definition
* *ssh*: SSH (sudo enabled) adapter

### Go version
The Go version employed has been go1.12.6 with modules enabled

## Do you think this is useful? back me up
Think and build this tool, has taken part of my time and effort. If you find it useful, and you think I deserve it, you can invite me a coffee :-)

<a href="https://www.buymeacoffee.com/jaumearus" target="_blank"><img src="https://cdn.buymeacoffee.com/buttons/default-orange.png" alt="Buy Me A Coffee" height="41" width="174"></a>
