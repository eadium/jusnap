
import glob
import os
import os.path
import re

from notebook.services.kernels.kernelmanager import MappingKernelManager


from notebook.prometheus.metrics import KERNEL_CURRENTLY_RUNNING_TOTAL
from notebook.utils import maybe_future

class ExternalIPythonKernelManager(MappingKernelManager):
    """A Kernel manager that connects to a IPython kernel started outside of Jupyter"""

    def _attach_to_persist_kernel(self, kernel_id, kid):
        self.log.info(f'Attaching {kernel_id} to an existing kernel...')
        kernel = self._kernels[kernel_id]
        port_names = ['shell_port', 'stdin_port', 'iopub_port', 'hb_port', 'control_port']
        for port_name in port_names:
            setattr(kernel, port_name, 0)

        # "Connect" to kernel started by an external python process
        connection_fname = f'{self.connection_dir}/kernel-{kid}.json'
        self.log.info(f'Latest kernel = {connection_fname} from dir = {self.connection_dir}')
        kernel.load_connection_file(connection_fname)

    async def start_kernel(self, **kwargs):
        if self._should_use_existing():
            kid = "persist"
        kernel_id = await super(ExternalIPythonKernelManager, self).start_kernel(**kwargs)
        self.log.info(f'Connection dir: {self.connection_dir}')
        self._attach_to_persist_kernel(kernel_id, kid)
        
        return kernel_id
