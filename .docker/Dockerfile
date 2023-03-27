FROM golang:1.20

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN make build

CMD ["make run-docker"]

#$ docker build -t my-golang-app .
#$ docker run -it --rm --name my-running-app my-golang-app
