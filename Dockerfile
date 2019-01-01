FROM golang:latest

RUN apt-get update && apt-get install -y --no-install-recommends \
		unzip \
	&& rm -rf /var/lib/apt/lists/*

WORKDIR /go/src/gitlab.com/disc/hlcup
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

EXPOSE 80

CMD ["make","run"]