PYTHON_VER = 3.9
PYTHON = python$(PYTHON_VER)
HOST=jusnap.rasseki.org

run: bin/jusnap conf.yml
	sudo --preserve-env ./bin/jusnap \
	--config conf.yml

clean:
	sudo rm -r dumps bin \
		python_modules/ipykernel/profile_default/db \
		python_modules/ipykernel/profile_default/log  \
		python_modules/ipykernel/profile_default/security \
		python_modules/ipykernel/profile_default/startup \
		python_modules/ipykernel/profile_default/pid \
		python_modules/ipykernel/profile_default/history.sqlite || true
	cd criu && $(MAKE) clean

install: build_criu jupyter go bin/jusnap

install_deb: criu_deb jupyter go bin/jusnap

jupyter:
	sudo add-apt-repository -y ppa:deadsnakes/ppa
	sudo apt install -y $(PYTHON) $(PYTHON)-distutils curl jupyter
	curl -sS https://bootstrap.pypa.io/get-pip.py | $(PYTHON)
	$(PYTHON) -m pip install --upgrade setuptools notebook pyzmq
	$(PYTHON) -m pip install -e ./python_modules/jupyter/extkern
	sed -i "s/example.com/$(HOST)/" ./python_modules/jupyter/jupyter_notebook_config.py
	jupyter nbextension install python_modules/jupyter/extensions/jusnap --user
	jupyter nbextension enable jusnap/jusnap

ipykernel_extension:
	mkdir -p ~/.ipython/extensions || true
	cp python_modules/ipykernel/extensions/snaphook.py ~/.ipython/extensions
	sed -i "s/# //" ./python_modules/ipykernel/profile_default/ipython_kernel_config.py

ipykernel_extension_off:
	sed -i "s/'snaphook'/# 'snaphook'/" ./python_modules/ipykernel/profile_default/ipython_kernel_config.py

go:
	sudo snap install go --classic
	go version

conf.yml:
	cp conf.yml.dist conf.yml

bin/jusnap:
	CGO=0 GOOS=linux go build -o ./bin/jusnap cmd/main.go

prep_build_criu:
	sudo apt update
	sudo apt install -y git curl build-essential libc6-dev-i386 gcc-multilib  pkg-config python-ipaddress \
		python3-future iproute2 libbsd-dev libcap-dev libnl-3-dev libnet-dev libaio-dev libprotobuf-dev \
		libprotobuf-c-dev protobuf-c-compiler protobuf-compiler python-protobuf

build_criu: prep_build_criu
	git clone 'https://github.com/checkpoint-restore/criu.git' || true
	sudo apt remove -y criu || true
	git config --global --add safe.directory "$$PWD"
	cd criu && $(MAKE) && sudo $(MAKE) install-lib install-criu install-compel install-amdgpu_plugin

criu_deb:
	sudo add-apt-repository -y ppa:criu/ppa
	sudo apt install -y criu
	curl --silent https://raw.githubusercontent.com/checkpoint-restore/criu/master/scripts/criu-ns > criu-ns
	chmod +x criu-ns
	sudo cp criu-ns /usr/local/sbin/

nginx:
	sudo apt install -y nginx certbot python3-certbot-nginx
	sed -i "s/example.com/$(HOST)/" ./nginx.conf
	sudo cp nginx.conf /etc/nginx/conf.d/jusnap.conf
	sudo rm /etc/nginx/sites-enabled/default || true
	sudo certbot --register-unsafely-without-email -n --keep-until-expiring --reuse-key --nginx -d $(HOST)
	sudo nginx -s reload

uninstall:
	$(MAKE) clean
	rm ~/.ipython/extensions/snaphook.py 
	jupyter nbextension disable jusnap/jusnap
	jupyter nbextension uninstall jusnap
	cd criu && sudo $(MAKE) uninstall
	sudo apt remove -y criu || true
	sudo apt autoremove