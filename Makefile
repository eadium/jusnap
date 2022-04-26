launch:
	sudo --preserve-env ./jusnap \
	--config conf.yml

clean:
	sudo rm -r dumps

sync:
	rsync -r -e 'ssh -p 7777' jusnap $$REMOTE:jusnap/
	rsync -r -e 'ssh -p 7777' Makefile $$REMOTE:jusnap/
	rsync -r -e 'ssh -p 7777' conf.yml $$REMOTE:jusnap/

build:
	CGO=0 GOOS=linux go build .

remote: build sync