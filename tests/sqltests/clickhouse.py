import json
from functools import cached_property
from uuid import UUID

import clickhouse_connect

from .types import TestCheck


class UUIDEncoder(json.JSONEncoder):
    def default(self, obj):
        if isinstance(obj, UUID):
            return obj.hex
        return json.JSONEncoder.default(self, obj)


class ClickHouseTestInstallation:
    # TODO: make context manager?
    def prepare(self) -> None:
        """Up ClickHouse instance with docker-compose"""

    def perform_check(self, check: TestCheck) -> None:
        query_result = self._client.query(check.query)
        json_rows = json.dumps(query_result.result_rows, cls=UUIDEncoder)
        result = json.loads(json_rows)

        if result != check.result:
            raise AssertionError(f'Expected {check.result}, got {result}')

    def cleanup(self) -> None:
        """Delete ClickHouse instance with docker-compose"""

    @cached_property
    def _client(self):
        return clickhouse_connect.get_client(
            host='localhost',
            port=8123,
            username='default',
            password='default',
        )
