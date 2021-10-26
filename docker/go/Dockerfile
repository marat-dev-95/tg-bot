FROM golang:1.17.2

ENV APP_NAME main
ENV PORT 8080

COPY ./ /go/src/${APP_NAME}
WORKDIR /go/src/${APP_NAME}

RUN go get ./cmd/tgbot
#RUN go build -o ${APP_NAME}

CMD ./${APP_NAME}

EXPOSE ${PORT}