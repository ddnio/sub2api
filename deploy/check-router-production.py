#!/usr/bin/env python3
"""Read-only production target check. Never prints Docker env or runs mutations."""
import json
import socket
import subprocess
import sys

HOST = "iZt4nfdwwipifuqc4q2v01Z"
ROOT = "/home/nio/recovery/live-production-20260906"
NAMES = ["fx-production-router", "fx-production-studio", "fx-production-postgres", "fx-production-redis"]


def validate(hostname, containers):
    errors = []
    if hostname != HOST:
        return ["wrong production hostname"]
    rows = {row.get("Name", "").lstrip("/"): row for row in containers}
    for name in NAMES:
        if name not in rows:
            errors.append(name + ": missing")
            continue
        row = rows[name]
        if not row.get("State", {}).get("Running"):
            errors.append(name + ": not running")
        nets = set(row.get("NetworkSettings", {}).get("Networks", {}))
        expected = {"fx-production-internal"}
        if name.endswith("router"):
            expected.add("fx-router-egress")
        if name.endswith("studio"):
            expected.add("fx-studio-egress")
        if nets != expected:
            errors.append(name + ": unexpected networks")
        mounts = row.get("Mounts", [])
        if name.endswith("router"):
            if not any(m.get("Type") == "bind" and m.get("Source") == ROOT + "/router.preview.json"
                       and m.get("Destination") == "/app/data/config.yaml" and m.get("RW") is False for m in mounts):
                errors.append(name + ": unexpected config mount")
        else:
            volume, dest = {
                "fx-production-studio": ("fx-production-studiodata", "/data"),
                "fx-production-postgres": ("fx-production-pgdata", "/var/lib/postgresql"),
                "fx-production-redis": ("fx-production-redisdata", "/data"),
            }[name]
            if not any(m.get("Type") == "volume" and m.get("Name") == volume and m.get("Destination") == dest for m in mounts):
                errors.append(name + ": unexpected data volume")
        ports = row.get("HostConfig", {}).get("PortBindings") or {}
        expected_ports = {}
        if name.endswith("router"):
            expected_ports = {"8080/tcp": [{"HostIp": "127.0.0.1", "HostPort": "18080"}]}
        if name.endswith("studio"):
            expected_ports = {"8788/tcp": [{"HostIp": "127.0.0.1", "HostPort": "18789"}]}
            env = dict(v.split("=", 1) for v in row.get("Config", {}).get("Env", []) if "=" in v)
            for key, value in {"STUDIO_PUBLIC_ORIGIN": "https://studio.nanafox.com", "ROUTER_AUTH_BASE_URL": "https://router.nanafox.com"}.items():
                if env.get(key) != value:
                    errors.append(name + ": unexpected " + key)
            if row.get("Config", {}).get("StopTimeout") != 600:
                errors.append(name + ": unexpected stop timeout")
        if ports != expected_ports:
            errors.append(name + ": unexpected published ports")
    return errors


def main():
    if socket.gethostname() != HOST:
        print("FAIL: wrong production hostname; Docker was not accessed", file=sys.stderr)
        return 1
    try:
        result = subprocess.run(["docker", "inspect", *NAMES], capture_output=True, text=True, timeout=30, check=True)
        errors = validate(socket.gethostname(), json.loads(result.stdout))
    except (OSError, ValueError, subprocess.SubprocessError):
        print("FAIL: cannot inspect required production containers", file=sys.stderr)
        return 1
    for error in errors:
        print("FAIL: " + error, file=sys.stderr)
    if errors:
        return 1
    print("PASS: Router production host, containers, ports, mounts, networks and Studio origin match")
    print("Read-only targeting check only; app health, DB connections and backup freshness still require verification")
    return 0


if __name__ == "__main__":
    sys.exit(main())
