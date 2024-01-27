import subprocess
import tempfile

from .terraform import TerraformLaunch
from .types import TestSuite, TestCase, TestCheck, TfConfigFile
from .clickhouse import ClickHouseTestInstallation
import shutil
import os


def run_tests(test_suite: TestSuite) -> None:
    chi = ClickHouseTestInstallation()
    chi.prepare()

    test_dir = tempfile.mkdtemp(prefix='terraform-provider-clickhouse-')
    tf = TerraformLaunch(test_dir)
    tf.init()

    for test in test_suite.tests:
        prepare_test(test, test_dir)
        tf.apply()
        for check in test.checks:
            chi.perform_check(check)

        clean_after_test(test, test_dir)

    shutil.rmtree(test_dir)
    chi.cleanup()


def prepare_test(test_case: TestCase, dest_dir: str):
    source_dir = os.path.dirname(__file__) + '/fixtures/'
    files = os.listdir(source_dir)

    for file_name in files:
        source_path = os.path.join(source_dir, file_name)
        destination_path = os.path.join(dest_dir, file_name)
        shutil.copy2(source_path, destination_path)

    for file_data in test_case.input:
        with open(os.path.join(dest_dir, file_data.name), 'w') as f:
            f.write(file_data.content)


def clean_after_test(test_case: TestCase, dest_dir: str):
    for file_data in test_case.input:
        os.remove(os.path.join(dest_dir, file_data.name))
