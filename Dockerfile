FROM golang:latest

RUN apt-get update && apt-get install -y --no-install-recommends \
		unzip \
	&& rm -rf /var/lib/apt/lists/*

RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

WORKDIR /go/src/github.com/disc/hlcup

COPY Gopkg.toml Gopkg.lock ./

RUN dep ensure --vendor-only

COPY . ./

RUN go build -o app

EXPOSE 80

CMD ["make","run"]