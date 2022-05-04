PYTHON_VER = 3.9
PYTHON = python$(PYTHON_VER)

launch: jusnap
	sudo --preserve-env ./bin/jusnap \
	--config conf.yml

clean:
	sudo rm -r dumps bin \
		~/.ipython/extensions/snaphook.py \
		python_modules/ipykernel/profile_default/db \
		python_modules/ipykernel/profile_default/log  \
		python_modules/ipykernel/profile_default/security \
		python_modules/ipykernel/profile_default/startup \
		python_modules/ipykernel/profile_default/pid \
		python_modules/ipykernel/profile_default/history.sqlite || true
	cd criu && $(MAKE) clean
sync:
	rsync -r -e 'ssh -p 7777' bin/jusnap $$REMOTE:jusnap/
	rsync -r -e 'ssh -p 7777' Makefile $$REMOTE:jusnap/
	rsync -r -e 'ssh -p 7777' conf.yml $$REMOTE:jusnap/

jusnap:
	CGO=0 GOOS=linux go build -o ./bin/jusnap cmd/main.go

remote: jusnap sync

install: prep_ubuntu go criu jupyter

prep_ubuntu:
	sudo cp /etc/apt/sources.list /etc/apt/sources.list~
	sudo sed -Ei 's/^# deb-src /deb-src /' /etc/apt/sources.list
	sudo add-apt-repository -y ppa:deadsnakes/ppa
	sudo apt update
	sudo apt build-dep -y criu
	sudo apt install -y git curl build-essential libc6-dev-i386 gcc-multilib  pkg-config python-ipaddress \
		python3-future iproute2 libbsd-dev libcap-dev libnl-3-dev libnet-dev libaio-dev libprotobuf-dev \
		libprotobuf-c-dev protobuf-c-compiler protobuf-compiler python-protobuf
	sudo apt install -y $(python)  $(python)-distutils
	curl -sS https://bootstrap.pypa.io/get-pip.py | $(python)

criu:
	git clone 'https://github.com/ccheckpoint-restore/criu.git'
	cd criu && make && sudo make install

jupyter:
	$(PYTHON) -m pip install --upgrade setuptools notebook
	$(PYTHON) -m pip install -e ./python_modules/jupyter/extkern
	$(PYTHON) -m jupyter nbextension install python_modules/jupyter/extensions/jusnap --user
	$(PYTHON) -m jupyter nbextension enable jusnap/jusnap

ipykernel_extension:
	mkdir -p ~/.ipython/extensions || true
	cp python_modules/ipykernel/extensions/snaphook.py ~/.ipython/extensions

go:
	sudo snap install go --classic
	go version
