launch:
	sudo --preserve-env ./jusnap

clean:
	sudo rm -r dumps

sync:
	rsync -r -e 'ssh -p 7777' jusnap ann.rasseki.org:jusnap/
	rsync -r -e 'ssh -p 7777' Makefile ann.rasseki.org:jusnap/

build:
	CGO=0 GOOS=linux go build .

remote: build sync