FROM golang

ADD . /go/src/github.com/austinov/go-recipes
WORKDIR /go/src/github.com/austinov/go-recipes

RUN curl https://glide.sh/get | sh
RUN glide up

RUN go test ./slack-bot/... -v
RUN go install ./slack-bot

ENTRYPOINT ["/go/bin/slack-bot", "-config", "./slack-bot/bot-dev.yaml"]
