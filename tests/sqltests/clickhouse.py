import json
import os
import random
import string
import subprocess
import time
from functools import cached_property
from uuid import UUID

import clickhouse_connect

from .types import TestCheck, TfChException


def generate_random_string(length: int) -> str:
    return ''.join(random.choice(string.ascii_lowercase) for _ in range(length))


class UUIDEncoder(json.JSONEncoder):
    def default(self, obj):
        if isinstance(obj, UUID):
            return obj.hex
        return json.JSONEncoder.default(self, obj)


class ClickHouseTestInstallation:
    def __init__(self, cwd: str, version: str = '23.12'):
        self.cwd = cwd
        self._env = {
            **os.environ,
            'CLICKHOUSE_LOCAL_PORT_HTTP': '18123',
            'CLICKHOUSE_LOCAL_PORT_NATIVE': '19000',
            'CLICKHOUSE_VERSION': version,
            'COMPOSE_PROJECT_NAME': f'tfch-{generate_random_string(8)}',
        }

    def __enter__(self):
        self.prepare()
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.cleanup()

    # TODO: Choose random port to allow parallel execution?
    def prepare(self) -> None:
        """Up ClickHouse instance with docker compose"""
        result = subprocess.run(
            ['docker', 'compose', 'up', '-d'],
            cwd=self.cwd,
            capture_output=True,
            text=True,
            env=self._env,
        )
        if result.returncode != 0:
            raise TfChException('ClickHouse initialization failed', result)

        self._check_clickhouse()

    def perform_check(self, check: TestCheck) -> None:
        query_result = self._client.query(check.query)
        json_rows = json.dumps(query_result.result_rows, cls=UUIDEncoder)
        result = json.loads(json_rows)

        if result != check.result:
            raise AssertionError(f'Expected {check.result}, got {result}')

    def cleanup(self) -> None:
        """Delete ClickHouse instance with docker compose"""
        result = subprocess.run(
            ['docker', 'compose', 'down'],
            cwd=self.cwd,
            capture_output=True,
            text=True,
            env=self._env,
        )
        if result.returncode != 0:
            raise TfChException('ClickHouse clean-up failed', result)

    @cached_property
    def _client(self):
        return clickhouse_connect.get_client(
            host='localhost',
            port=18123,
            username='default',
            password='default',
        )

    def _check_clickhouse(self) -> None:
        max_attempts = 10
        attempts = 0
        while attempts < max_attempts:
            try:
                self._client.ping()
                break
            except Exception:
                attempts += 1
                time.sleep(1)
        else:
            raise TfChException('Cannot connect to ClickHouse instance', self)
