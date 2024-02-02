import os
import subprocess

from .types import TfChException


class Terraform:
    def __init__(self, cwd):
        self.cwd = cwd

    def init(self):
        result = subprocess.run(
            ['terraform', 'init'],
            cwd=self.cwd,
            capture_output=True,
            text=True,
        )
        if result.returncode != 0:
            raise TfChException('terraform init failed', result)

    def apply(self):
        result = subprocess.run(
            ['terraform', 'apply', '-auto-approve', '-no-color'],
            cwd=self.cwd,
            capture_output=True,
            text=True,
            env={**os.environ, 'TF_LOG': 'debug'}
        )
        if result.returncode != 0:
            raise TfChException('terraform apply failed', result)
