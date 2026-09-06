import copy
import importlib.util
import io
import json
from pathlib import Path
import subprocess
import tempfile
import unittest
from unittest.mock import patch
from contextlib import redirect_stderr

ROOT = Path(__file__).resolve().parents[2]
spec = importlib.util.spec_from_file_location("preflight", ROOT / "deploy/check-router-production.py")
check = importlib.util.module_from_spec(spec)
spec.loader.exec_module(check)


def fixture():
    rows = []
    for name in check.NAMES:
        nets = {"fx-production-internal": {}}
        mounts = []
        ports = {}
        config = {}
        if name.endswith("router"):
            nets["fx-router-egress"] = {}
            mounts = [{"Type": "bind", "Source": check.ROOT + "/router.preview.json", "Destination": "/app/data/config.yaml", "RW": False}]
            ports = {"8080/tcp": [{"HostIp": "127.0.0.1", "HostPort": "18080"}]}
        else:
            volume, dest = {"fx-production-studio": ("fx-production-studiodata", "/data"), "fx-production-postgres": ("fx-production-pgdata", "/var/lib/postgresql"), "fx-production-redis": ("fx-production-redisdata", "/data")}[name]
            mounts = [{"Type": "volume", "Name": volume, "Destination": dest}]
        if name.endswith("studio"):
            nets["fx-studio-egress"] = {}
            ports = {"8788/tcp": [{"HostIp": "127.0.0.1", "HostPort": "18789"}]}
            config = {"Env": ["STUDIO_PUBLIC_ORIGIN=https://studio.nanafox.com", "ROUTER_AUTH_BASE_URL=https://router.nanafox.com"], "StopTimeout": 600}
        rows.append({"Name": "/" + name, "State": {"Running": True}, "NetworkSettings": {"Networks": nets}, "Mounts": mounts, "HostConfig": {"PortBindings": ports}, "Config": config})
    return rows


class ProductionTargetTests(unittest.TestCase):
    def test_expected_layout(self):
        self.assertEqual(check.validate(check.HOST, fixture()), [])

    def test_wrong_host_never_accesses_docker(self):
        with patch.object(check.socket, "gethostname", return_value="old-host"), patch.object(check.subprocess, "run") as run, redirect_stderr(io.StringIO()):
            self.assertEqual(check.main(), 1)
            run.assert_not_called()

    def test_rejects_old_database_volume(self):
        rows = fixture(); rows[2]["Mounts"][0]["Name"] = "deploy_postgres_data"
        self.assertIn("fx-production-postgres: unexpected data volume", check.validate(check.HOST, rows))

    def test_rejects_old_config(self):
        rows = fixture(); rows[0]["Mounts"][0]["Source"] = "/etc/sub2api/prod.yaml"
        self.assertIn("fx-production-router: unexpected config mount", check.validate(check.HOST, rows))

    def test_rejects_missing_egress(self):
        rows = fixture(); del rows[0]["NetworkSettings"]["Networks"]["fx-router-egress"]
        self.assertTrue(check.validate(check.HOST, rows))

    def test_rejects_public_database_port(self):
        rows = fixture(); rows[2]["HostConfig"]["PortBindings"] = {"5432/tcp": [{"HostIp": "0.0.0.0", "HostPort": "5432"}]}
        self.assertTrue(check.validate(check.HOST, rows))

    def test_rejects_old_app_port(self):
        rows = fixture(); rows[0]["HostConfig"]["PortBindings"]["8080/tcp"][0]["HostPort"] = "8080"
        self.assertTrue(check.validate(check.HOST, rows))

    def test_rejects_missing_or_stopped_container(self):
        self.assertTrue(check.validate(check.HOST, fixture()[:-1]))
        rows = fixture(); rows[0]["State"]["Running"] = False
        self.assertTrue(check.validate(check.HOST, rows))

    def test_origin_failure_does_not_print_value(self):
        rows = fixture(); rows[1]["Config"]["Env"][0] = "STUDIO_PUBLIC_ORIGIN=secret-sentinel"
        errors = check.validate(check.HOST, rows)
        self.assertTrue(errors)
        self.assertNotIn("secret-sentinel", str(errors))

    def test_legacy_prod_exits_before_tools(self):
        # Any accidental git/docker invocation is an observable failure.
        import os
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp); marker = root / "called"
            for tool in ("git", "docker", "hostname", "mkdir"):
                p = root / tool
                p.write_text("#!/bin/sh\n: > '" + str(marker) + "'\nexit 99\n")
                p.chmod(0o755)
            result = subprocess.run(["/bin/bash", str(ROOT / "deploy/deploy-server.sh"), "prod"], env={**os.environ, "PATH": tmp}, capture_output=True)
            self.assertEqual(result.returncode, 64)
            self.assertFalse(marker.exists())

    def test_legacy_test_refuses_production_host(self):
        import os
        with tempfile.TemporaryDirectory() as tmp:
            p = Path(tmp) / "hostname"; p.write_text("#!/bin/sh\necho " + check.HOST + "\n"); p.chmod(0o755)
            result = subprocess.run(["/bin/bash", str(ROOT / "deploy/deploy-server.sh"), "test"], env={**os.environ, "PATH": tmp}, capture_output=True)
            self.assertEqual(result.returncode, 64)


if __name__ == "__main__":
    unittest.main()
