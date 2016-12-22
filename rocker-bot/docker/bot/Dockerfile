FROM golang

ADD . /go/src/github.com/austinov/go-recipes
WORKDIR /go/src/github.com/austinov/go-recipes

RUN curl https://glide.sh/get | sh
RUN glide up

RUN go test ./rocker-bot/... -v
RUN go install ./rocker-bot

ENTRYPOINT ["/go/bin/rocker-bot", "-config", "./rocker-bot/bot-dev.yaml"]
