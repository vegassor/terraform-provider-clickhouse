import pathlib
from typing import Generator

from sqltests import YamlFile


def pytest_collect_file(parent, file_path: pathlib.PosixPath) -> Generator[YamlFile, None, None]:
    if file_path.suffix == '.yaml' and not file_path.name.startswith('docker'):
        return YamlFile.from_parent(parent, path=file_path)
