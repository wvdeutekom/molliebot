FROM alpine:3.5

ENV GOROOT=/usr/lib/go \
    GOPATH=/gopath \
    GOBIN=/gopath/bin \
    PATH=$PATH:$GOROOT/bin:$GOPATH/bin

WORKDIR /gopath/src/app
ADD . /gopath/src/app

# Install go and dependencies
# Cleanup afterwards
RUN apk add -U git go musl-dev && \
  go get -v app && \
  apk del git go && \
  rm -rf /gopath/pkg && \
  rm -rf /gopath/src && \
  rm -rf /var/cache/apk/*

# We need this after the cleanup for the bot to make a websocket connection to slack
RUN apk add -U ca-certificates

# Set the right timezone
RUN apk add tzdata && \
  cp /usr/share/zoneinfo/Europe/Amsterdam /etc/localtime && \
  echo "Europe/Amsterdam" >  /etc/timezone && \
  apk del tzdata

CMD ["/gopath/bin/app"]