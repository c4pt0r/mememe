# syntax=docker/dockerfile:1
FROM golang:1.18-buster
WORKDIR /app

RUN apt install git
RUN git clone https://github.com/c4pt0r/mememe .
RUN go build -o /bin/mememe

CMD [ "/bin/mememe" ]

ENTRYPOINT /bin/mememe -tgbot-token ${TGBOT_TOKEN}
