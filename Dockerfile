FROM golang:latest

RUN apt-get update && apt-get install -y --no-install-recommends \
		unzip \
	&& rm -rf /var/lib/apt/lists/*

RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

WORKDIR /go/src/gitlab.com/disc/hlcup
COPY . .

RUN dep ensure
RUN go install -v ./...

EXPOSE 80

CMD ["make","run"]