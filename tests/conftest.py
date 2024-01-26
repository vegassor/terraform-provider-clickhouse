import pytest, yaml
from typing import Any

from pydantic import BaseModel

class TfConfigFile(BaseModel):
    name: str
    content: str

class TestCheck(BaseModel):
    sql: str
    result: Any

class TestCase(BaseModel):
    name: str
    input: list[TfConfigFile]
    checks: list[TestCheck]

class TestSuite(BaseModel):
    tests: list[TestCase]

def pytest_collect_file(parent, file_path):
    if file_path.suffix == ".yaml":
        return YamlFile.from_parent(parent, path=file_path)


class YamlFile(pytest.File):
    def collect(self):
        raw = yaml.safe_load_all(self.path.open(encoding="utf-8"))
        spec = TestSuite(tests=raw)
        yield YamlItem.from_parent(self, name=self.path.name, spec=spec)


class YamlItem(pytest.Item):
    def __init__(self, *, spec: TestSuite, **kwargs):
        super().__init__(**kwargs)
        self.ts = spec

    def runtest(self) -> None:
        if self.ts.tests[0].checks[0].result != ['mydb']:
            raise TfChException(self, 'TODO', 'todo')

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


class TfChException(Exception):
    """Custom exception for error reporting."""