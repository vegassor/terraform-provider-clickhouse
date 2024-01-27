import os
import shutil
import tempfile

from .clickhouse import ClickHouseTestInstallation
from .terraform import Terraform
from .types import TestSuite, TestCase, TestCheck, TfConfigFile


def run_tests(test_suite: TestSuite) -> None:
    test_dir = tempfile.mkdtemp(prefix='terraform-provider-clickhouse-')
    source_dir = f'{os.path.dirname(__file__)}/fixtures/'
    shutil.copytree(source_dir, test_dir,  dirs_exist_ok=True)

    chi = ClickHouseTestInstallation(f'{test_dir}/clickhouse')
    chi.prepare()

    tf = Terraform(test_dir)
    tf.init()

    for test in test_suite.tests:
        prepare_test(test, test_dir)
        tf.apply()
        for check in test.checks:
            chi.perform_check(check)

        clean_after_test(test, test_dir)

    chi.cleanup()
    shutil.rmtree(test_dir)


def prepare_test(test_case: TestCase, dest_dir: str):
    for file_data in test_case.input:
        with open(os.path.join(dest_dir, file_data.name), 'w') as f:
            f.write(file_data.content)


def clean_after_test(test_case: TestCase, dest_dir: str):
    for file_data in test_case.input:
        os.remove(os.path.join(dest_dir, file_data.name))
