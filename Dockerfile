FROM golang:bullseye

WORKDIR /home/xonstat

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

RUN apt-get update -y && apt-get install -y libcairo2-dev

COPY . .
RUN go build -trimpath

CMD ["xonstat"]
