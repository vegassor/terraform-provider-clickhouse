import dataclasses
import os
import shutil
import tempfile

import pytest
import yaml

from .clickhouse import ClickHouseTestInstallation
from .terraform import Terraform
from .types import TestSuite, TestCase, TestCheck, TfConfigFile


@dataclasses.dataclass(frozen=True)
class Context:
    terraform: Terraform
    clickhouse: ClickHouseTestInstallation
    cwd: str


class YamlFile(pytest.File):
    def collect(self):
        raw = yaml.safe_load_all(self.path.open(encoding="utf-8"))
        suite = TestSuite(name=self.path.name, testcases=raw)
        yield YamlItem.from_parent(
            self,
            name=suite.name,
            spec=suite,
        )


class YamlItem(pytest.Item):
    def __init__(self, *, spec: TestSuite, **kwargs):
        super().__init__(**kwargs)
        self.suite = spec
        self.cwd = tempfile.mkdtemp(prefix='terraform-provider-clickhouse-')

    def runtest(self) -> None:
        source_dir = f'{os.path.dirname(__file__)}/fixtures/'
        shutil.copytree(source_dir, self.cwd,  dirs_exist_ok=True)

        chi = ClickHouseTestInstallation(f'{self.cwd}/clickhouse')
        chi.prepare()

        tf = Terraform(self.cwd)
        tf.init()

        for testcase in self.suite.testcases:
            self._prepare_test(testcase)

            tf.apply()
            for check in testcase.checks:
                chi.perform_check(check)

            self._clean_after_test(testcase)

        chi.cleanup()
        shutil.rmtree(self.cwd)

    def _prepare_test(self, case: TestCase):
        for file_data in case.input:
            with open(os.path.join(self.cwd, file_data.name), 'w') as f:
                f.write(file_data.content)

    def _clean_after_test(self, case: TestCase):
        for file_data in case.input:
            os.remove(os.path.join(self.cwd, file_data.name))
