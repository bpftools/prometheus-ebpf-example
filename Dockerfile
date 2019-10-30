FROM golang:1.13-alpine

RUN sed -i -e 's/v[[:digit:]]\..*\//edge\//g' /etc/apk/repositories

RUN apk update
RUN apk add bcc-dev build-base linux-headers

COPY . /project
WORKDIR /project
RUN go build -o /bin/bpf-program

RUN rm -rf /project

ENTRYPOINT ["/bin/bpf-program"]
