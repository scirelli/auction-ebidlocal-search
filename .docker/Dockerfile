FROM golang:1.20

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN apt-get update && apt-get install -y jq && go mod download && go mod verify

COPY . .
RUN make build

ENTRYPOINT ["make", "run-docker"]
CMD ["make", "run-docker", "configPath=/usr/src/app/configs/config-docker.json"]

#$ mkdir -p data_dir/web/user data_dir/web/watchlists && ./.docker/docker-volume-cp data_dir/. ebidUserData:.
#$ docker build --tag ebidlocal-watchlist --file ./.docker/Dockerfile .
#$ docker run --interactive --tty --rm --publish 8282:80 -v ebidUserData:/data --name ebidlocal-server ebidlocal-watchlist
