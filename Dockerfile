# FROM --platform=linux/amd64 golang:1.17.12-alpine3.16 as builder

# ENV LANG en_US.utf8
# ENV ARCH amd64
# ENV GOVERSION 1.17.12
# ENV ROOT /app
# ENV GOPATH /go
# ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
# ENV DEBIAN_FRONTEND=noninteractive

# WORKDIR $ROOT
# COPY . $ROOT

# RUN apk add build-base
# RUN go mod download
# RUN mkdir -p /tmp/log/app
# RUN mkdir -p /tmp/app

# EXPOSE 80

# ENTRYPOINT go build -o app ./main.go  && ./app


FROM --platform=linux/amd64 ubuntu:21.10 as app

ENV LANG en_US.utf8
ENV ARCH amd64
ENV GOVERSION 1.7.5
ENV ROOT /app
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
ENV DEBIAN_FRONTEND=noninteractive

ENV END=prod \
    REDIS_HOST=tdt-broadcast.v81cls.ng.0001.apne1.cache.amazonaws.com \
    REDIS_PORT=6379

RUN apt-get -y update \
    && apt-get -y upgrade \
    && apt -y update \
    && apt -y upgrade

RUN apt install -y apt-transport-https
RUN apt install -y ca-certificates
RUN apt install -y curl
RUN apt install -y software-properties-common

RUN apt-get install -y wget
RUN apt-get install -y git
RUN apt-get install -y gcc
RUN apt install -y golang-go

RUN apt-get install -y locales curl
RUN apt-get install -y openssh-server
RUN apt -y install mysql-server
RUN apt -y install mysql-client
RUN apt -y install libmysqlclient-dev
RUN apt -y install redis-server

# RUN curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
# RUN add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu focal stable"
# RUN apt update

RUN echo 'root:root' | chpasswd
RUN mkdir /run/sshd
RUN sed -i 's/#PasswordAuthentication yes/PasswordAuthentication yes/' /etc/ssh/sshd_config
RUN sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config

EXPOSE 22
EXPOSE 80

WORKDIR $ROOT
COPY . $ROOT

# ADD start.sh ./start.sh
# RUN chmod -x ./start.sh
# CMD ["./start.sh"]
CMD ["/usr/sbin/sshd", "-D"]
