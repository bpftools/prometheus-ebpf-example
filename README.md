# Prometheus eBPF example

A simple `main.go` containing all the code you need to get metrics from the kernel and expose them trough [Prometheus](https://prometheus.io/).


## Build

You need [Docker](https://docs.docker.com/install/) to build this using the makefile.

```bash
make build
```

If you don't want to use Docker, and the Makefile you can build locally with:


```bash
go build -o bpf-program .
```

In this case, you will need to install bcc-dev first, instructions [here](https://github.com/iovisor/bcc/blob/master/INSTALL.md).

## Run

Remember: you might have libreadline somewhere else. This also can work by passing the bash binary directly. If you pass the bash binary or change
the location you can use the environment variable `URETPROBE_BINARY` to tell the program where to load from.

```bash
docker run -e NODENAME=$(hostname) -v /sys:/sys:ro -v /lib/modules:/lib/modules:ro --privileged -v /lib/libreadline.so:/host/lib/libreadline.so:ro -p 8080:8080 -it docker.io/bpftools/prometheus-ebpf-example:latest
```

You can test if this works by opening a `bash` shell and doing some commands, then you can curl
the metrics endpoint and see the results. It will show something like:

```bash
# HELP commands_count The number of times a command is invoked via bash
# TYPE commands_count counter
commands_count{command="clear",nodename="gallifrey",pid="1834654"} 3
commands_count{command="curl http://127.0.0.1:8080/metrics",nodename="gallifrey",pid="1847919"} 1
commands_count{command="docker images",nodename="gallifrey",pid="1834654"} 1
commands_count{command="docker ps",nodename="gallifrey",pid="1834654"} 1
commands_count{command="ip a",nodename="gallifrey",pid="1834654"} 1
commands_count{command="ip a",nodename="gallifrey",pid="1847919"} 2
commands_count{command="ls -la",nodename="gallifrey",pid="1834654"} 1
commands_count{command="ls -la",nodename="gallifrey",pid="1847919"} 4
commands_count{command="ps",nodename="gallifrey",pid="1834654"} 1
commands_count{command="ps -fe",nodename="gallifrey",pid="1834654"} 1
commands_count{command="ps -fe | grep evil",nodename="gallifrey",pid="1834654"} 1
commands_count{command="vim",nodename="gallifrey",pid="1834654"} 1
commands_count{command="vim",nodename="gallifrey",pid="1847919"} 2
commands_count{command="whoami",nodename="gallifrey",pid="1834654"} 1
```

Notice how the curl command itself was recorded!

## Run on Kubernetes as a Daemonset

Remember: you might have libreadline somewhere else. This also can work by passing the bash binary directly. If you pass the bash binary or change
the location you can use the environment variable `URETPROBE_BINARY` to tell the program where to load from.
In this case, the `daemonset.yaml` uses the `URETPROBE_BINARY` because in my testing I was using [kind](https://github.com/kubernetes-sigs/kind) and it worked well in that way.


```bash
kubectl apply -f daemonset.yaml
```

This will create:
- A namespace called bpf-stuff
- A daemonset called bpf-program
- A service exposing port 8080 called bpf-program


At this point you can port-forward your daemonset or the service or even one of the pods
and curl the endpoint to see if it's collecting anything:

To do the port forwarding for the daemonset:

```bash
kubectl port-forward daemonset/bpf-program -n bpf-stuff 8080:8080
```


## Scrape with Prometheus

TODO: Write this

## Licensing

Since the `main.go` file in this repository uses the original work taken from the [bash_readline.go](https://github.com/iovisor/gobpf/blob/master/examples/bcc/bash_readline/bash_readline.go)
example in the [iovisor/gobpf](https://github.com/iovisor/gobpf) repository, we use the same license as them.
Modifications are stated in the `main.go` file.
