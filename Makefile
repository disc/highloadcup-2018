build:
	docker build -t hlcup .
run-local: build
	@docker rm -f $$(docker ps -qa -f name=hlcup) || true
	docker run --name hlcup --rm -p 8080:80 -v $$(pwd)/data.zip:/tmp/data/data.zip -v $$(pwd)/data/options.txt:/go/src/github.com/disc/hlcup/data/options.txt -t hlcup
deploy: build
	docker tag hlcup stor.highloadcup.ru/accounts/rebel_butterfly
	docker push stor.highloadcup.ru/accounts/rebel_butterfly

run: app-unzip app-use-options
	./app
app-unzip:
	mkdir -p $$(pwd)/data/ > /dev/null
	unzip -oq /tmp/data/data.zip -d $$(pwd)/data/
app-use-options:
	if [ -e /tmp/data/options.txt ] ; \
    then \
         cp /tmp/data/options.txt $$(pwd)/data/ > /dev/null ; \
    fi;
tester:
	highloadcup_tester -addr http://127.0.0.1:8080 -hlcupdocs ~/Downloads/test_accounts_291218 -test -phase 1