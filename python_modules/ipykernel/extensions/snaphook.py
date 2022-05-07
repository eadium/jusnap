import requests


class CellWatcher(object):
    def __init__(self, ip):
        self.shell = ip
        self.url = "http://localhost:8000/api/snap/new"

    def post_run_cell(self, result):
        x = requests.post(self.url, data={}, timeout=5)
        print(x.text)


def load_ipython_extension(ip):
    vw = CellWatcher(ip)
    ip.events.register("post_run_cell", vw.post_run_cell)
