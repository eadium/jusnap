import extkern

c.NotebookApp.kernel_manager_class = extkern.ExternalIPythonKernelManager
c.KernelManager.autorestart = False
c.Session.key = b""
c.NotebookApp.open_browser = False
c.NotebookApp.allow_origin = 'https://example.com'