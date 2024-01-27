import pathlib
from typing import Generator

import pytest
import yaml

from sqltests import TestSuite, run_tests


class YamlFile(pytest.File):
    def collect(self):
        raw = yaml.safe_load_all(self.path.open(encoding="utf-8"))
        spec = TestSuite(name=self.path.name, tests=raw)
        yield YamlItem.from_parent(self, name=spec.name, spec=spec)


class YamlItem(pytest.Item):
    def __init__(self, *, spec: TestSuite, **kwargs):
        super().__init__(**kwargs)
        self.ts = spec

    def runtest(self) -> None:
        run_tests(self.ts)

    def repr_failure(self, excinfo):
        """Called when self.runtest() raises an exception."""
        if isinstance(excinfo.value, TfChException):
            _, expected, got = excinfo.value.args
            return (
                f"usecase execution failed"
                f"   spec failed: {expected!r} != {got!r}"
                f"   no further details known at this point."
            )
        return super().repr_failure(excinfo)

    def reportinfo(self):
        return self.path, 0, f"usecase: {self.name}"


def pytest_collect_file(parent, file_path: pathlib.PosixPath) -> Generator[YamlFile, None, None]:
    if file_path.suffix == '.yaml' and not file_path.name.startswith('docker'):
        return YamlFile.from_parent(parent, path=file_path)


class TfChException(Exception):
    pass
