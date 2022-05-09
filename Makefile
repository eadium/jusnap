PYTHON_VER = 3.9
PYTHON = python$(PYTHON_VER)

launch: bin/jusnap conf.yml
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
	jupyter nbextension disable jusnap/jusnap
	jupyter nbextension uninstall jusnap

bin/jusnap:
	CGO=0 GOOS=linux go build -o ./bin/jusnap cmd/main.go

install: build_criu jupyter go bin/jusnap

criu_deb:
	sudo add-apt-repository -y ppa:criu/ppa
	sudo apt install -y criu 

jupyter:
	sudo add-apt-repository -y ppa:deadsnakes/ppa
	sudo apt install -y $(PYTHON) $(PYTHON)-distutils curl jupyter
	curl -sS https://bootstrap.pypa.io/get-pip.py | $(PYTHON)
	$(PYTHON) -m pip install --upgrade setuptools notebook pyzmq
	$(PYTHON) -m pip install -e ./python_modules/jupyter/extkern
	jupyter nbextension install python_modules/jupyter/extensions/jusnap --user
	jupyter nbextension enable jusnap/jusnap

ipykernel_extension:
	mkdir -p ~/.ipython/extensions || true
	cp python_modules/ipykernel/extensions/snaphook.py ~/.ipython/extensions

go:
	sudo snap install go --classic
	go version

conf.yml:
	cp conf.yml.dist conf.yml

prep_build_criu:
	# sudo cp /etc/apt/sources.list /etc/apt/sources.list~
	# sudo sed -Ei 's/^# deb-src /deb-src /' /etc/apt/sources.list
	# sudo add-apt-repository -y ppa:criu/ppa
	sudo apt update
	# sudo apt build-dep -y criu
	sudo apt install -y git curl build-essential libc6-dev-i386 gcc-multilib  pkg-config python-ipaddress \
		python3-future iproute2 libbsd-dev libcap-dev libnl-3-dev libnet-dev libaio-dev libprotobuf-dev \
		libprotobuf-c-dev protobuf-c-compiler protobuf-compiler python-protobuf 

build_criu: prep_build_criu
	git clone 'https://github.com/checkpoint-restore/criu.git' || true
	sudo apt remove -y criu || true
	git config --global --add safe.directory "$$PWD"
	cd criu && $(MAKE) && sudo $(MAKE) install-lib install-criu install-compel install-amdgpu_plugin
