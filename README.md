# JuSnap - Jupyter Notebook snapshotting tool

## Purpose

Jusnap could be used for creating state-checkpoints (to be not confused with built-in Jupyter checkpoints, therefore called *snapshots*). Snapshots are useful in case of long-running calculation-heavy Jupyter operations, allowing you to store different states of variables and the whole kernel.

## Compatibility
Works fine for Ubuntu and other linux-based distros. Doesn't fit for Windows, Solaris and probably MacOS.

## Prerequisites

To use `make` you'll need GNU Make installed on your system
```bash
sudo apt update && sudo apt install make
```
### Installation

There are predefined installation steps for Ubuntu, tested on Ubuntu Server 18.04/20.04.
For other linux distributions, you can perform installation manually, guided by the contents of the Makefile.
Installation should be run by user eligible for `sudo` command, strictly **non-root**!

```bash
git clone https://github.com/eadium/jusnap.git
cd jusnap
make install
```

During installation in your system will be installed python3.9, golang, jupyter-notebook with ipykernel and CRIU alongside with lots of packages required to build CRIU from sources. This is the best option as CRIU will be built according to your system's environment.

However, if you want to install CRIU from deb package, please run:
```bash 
make install_deb
```

You may also want to use https with nginx. To do so type:
```bash
HOST=example.com make nginx
```
where `example.com` is your hostname (without scheme).
This command will install nginx and launch certbot to issue a certificate for your domain.


### Configuration
Jusnap could be configured either via CLI parameters or using a config file `conf.yml`. The latter is a preferable way.
Checkout an example in [conf.yml.dist](conf.yml.dist).

```
./jusnap -h
Usage of default:
Usage of default:
      --config string                              Configuration file path.
      --graceful.shutdown_timeout duration         Graceful shutdown timeout (default 10s)
      --jusnap.criu.ghost_limit int                Ghost file limit (MB) (default 2)
      --jusnap.http.port string                    HTTP port (default "8000")
      --jusnap.http.read_timeout duration          HTTP read timeout (default 5s)
      --jusnap.http.write_timeout duration         HTTP write timeout (default 5s)
      --jusnap.ipython.args strings                Launch arguments fot ipykernel
      --jusnap.ipython.cooldown duration           Snapshotting cooldown interval (default 5s)
      --jusnap.ipython.history_enabled             Enables history file management (default false)
      --jusnap.ipython.history_file string         Path to history.sqlite (default "~/.ipython/profile_default/history.sqlite")
      --jusnap.ipython.python_interpreter string   Python interpreter to use (default "python3")
      --jusnap.ipython.runtime_path string         Path to Jupyter runtime dir (default "~/.local/share/jupyter/runtime/")
      --jusnap.jupyter.args strings                Launch arguments fot Jupyter Notebook
      --jusnap.jupyter.port int                    TCP port for Jupyter Notebook (default 8888)
      --jusnap.log_level string                    Logging level (default "info")
      --jusnap.os.gid int                          GID for created files (default [running user gid])
      --jusnap.os.uid int                          UID for created files (default [running user uid])
```
Please pay attention to `--jusnap.ipython.args` and `--jusnap.jupyter.args` as these options allow you to pass extra arguments to ipython kernel and Jupyter server accordingly.

More info:
> [IPython CLI options](https://ipython.readthedocs.io/en/stable/config/options/terminal.html)

> [Jupyter Notebook CLI options](https://jupyter-notebook.readthedocs.io/en/stable/config.html)

Jusnap provides an auto-snapshotting feature, that makes IPython kernel snapshot after every executed cell, if cooldown period (`--jusnap.ipython.cooldown`) has passed. To enable it, please run:
```bash
make ipykernel_extension 
```
and to disable:
```bash
make ipykernel_extension_off
```
## Usage
The simplest way to start Jusnap is:
```bash
make
```
This command will run Jusnap with `conf.yml` as configuration file.
Jusnap will start a listening HTTP server using port from `--jusnap.http.port` and Jupyter Notebook server on default port 8888 (or any other user-supplied through `--jusnap.jupyter.port` config option)

Jupyter Notebook Output is logged to `jupyter.log` file.

## Inside
Jusnap is powered bu [CRIU](https://criu.org/Main_Page) snapshotting tool and consists of 4 main parts:
- golang server (API + process management)
- Jupyter kernel [manager](python_modules/jupyter/extkern/extkern/__init__.py)
- IPython [extension](python_modules/ipykernel/extensions/snaphook.py) for autosnapshotting
- Jupyter Notebook frontend [extension](python_modules/jupyter/extensions/jusnap/jusnap.js)


## TODOs
- Add network storage for images
- Fix support of history file (currently unsupported feature)
- Add HTTPS (that will also fix browser notifications)
- Add CentOS support for Makefile
- Add Jupyter Lab plugin