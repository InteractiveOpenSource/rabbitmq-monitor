# rabbitmq-monitor
CLI Tool to monitor RabbitMQ server - built with [Go](https://golang.org/)

It's a very basic tool to monitor a RabbitMQ Server by reading from API provided by your server

All you have to do is to build it and run it by specifying your server config

### Deployment
```
$ ./deploy.sh --build=/path/to/bin/rabbitmq-monitor
```

### Usage
```
$ rabbitmq-monitor monitor --host <host|default="localhost"> --port <port|default="15672"> --user <user> --password <password> --vhost <a-specific-vhost> --tick <in_milliseconds>
```