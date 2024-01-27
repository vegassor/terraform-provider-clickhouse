import subprocess
from .types import TfChException


class TerraformLaunch:
    def __init__(self, test_dir):
        self.test_dir = test_dir

    def init(self):
        result = subprocess.run(
            ['terraform', 'init'],
            cwd=self.test_dir,
            capture_output=True,
            text=True,
        )
        if result.returncode != 0:
            raise TfChException('terraform init failed', result)

    def apply(self):
        result = subprocess.run(
            ['terraform', 'apply', '-auto-approve'],
            cwd=self.test_dir,
            capture_output=True,
            text=True,
        )
        if result.returncode != 0:
            raise TfChException('terraform apply failed', result)  # TODO: add more info
