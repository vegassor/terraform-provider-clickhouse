from typing import Any
from pydantic import BaseModel


class TfConfigFile(BaseModel):
    name: str
    content: str


class TestCheck(BaseModel):
    query: str
    result: Any


class TestCase(BaseModel):
    name: str
    input: list[TfConfigFile]
    checks: list[TestCheck]


class TestSuite(BaseModel):
    name: str
    testcases: list[TestCase]


class TfChException(Exception):
    pass
