import requests

class CellWatcher(object):
    def __init__(self):
        self.url = "http://localhost:8000/api/snap/new"

    def post_execute(self):
        x = requests.post(self.url, data={}, timeout=5)
        print(x.text)

def load_ipython_extension(ip):
    vw = CellWatcher()
    ip.events.register("post_execute", vw.post_execute)
