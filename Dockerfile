FROM golang:1.14.4-buster as base_env

WORKDIR /app

COPY . .

RUN go mod vendor \
    && cd vendor\gocv.io\x\gocv \
    && make install \
    && cd - \
    && go build -o main