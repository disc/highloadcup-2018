build:
	docker build -t hlcup .
run-local: build
	@docker rm -f $$(docker ps -qa -f name=hlcup) || true
	docker run --name hlcup --rm -p 8080:80 -v $$(pwd)/data.zip:/tmp/data/data.zip -t hlcup
deploy: build
	docker tag hlcup stor.highloadcup.ru/accounts/rebel_butterfly
	docker push stor.highloadcup.ru/accounts/rebel_butterfly

run: app-unzip
	/go/bin/hlcup
app-unzip:
	mkdir -p $$(pwd)/data/ > /dev/null
	unzip -oq /tmp/data/data.zip -d $$(pwd)/data/
app-use-options:
	if [ -e /tmp/data/options.txt ] ; \
    then \
         cp /tmp/data/options.txt $$(pwd)/data/ > /dev/null ; \
    fi;