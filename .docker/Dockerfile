FROM golang:1.20

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN make build

CMD ["make run-docker"]

#$ docker build -t ebidlocal-watchlist ..
#$ docker run -it --rm --publish 8282:8282 --name ebidlocal-server ebidlocal-watchlist
