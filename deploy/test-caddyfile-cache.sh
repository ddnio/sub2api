#!/bin/sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)

for caddyfile in \
	"$repo_root/deploy/Caddyfile.server" \
	"$repo_root/deploy/Caddyfile.tob" \
	"$repo_root/deploy/Caddyfile.template"
do
	active_config=$(sed 's/[[:space:]]*#.*$//' "$caddyfile")
	if printf '%s\n' "$active_config" | grep -Eiq 'Cache-Control.*immutable'; then
		echo "$caddyfile must not force immutable caching; the backend owns asset cache policy" >&2
		exit 1
	fi
	if ! printf '%s\n' "$active_config" | grep -Eq '^[[:space:]]*reverse_proxy[[:space:]]+(localhost|127\.0\.0\.1):808[01]'; then
		echo "$caddyfile must continue proxying application routes to the backend" >&2
		exit 1
	fi
	if ! printf '%s\n' "$active_config" | grep -Eq '^[[:space:]]*max_header_size[[:space:]]+64KB[[:space:]]*$'; then
		echo "$caddyfile must keep the 64 KiB request-header limit" >&2
		exit 1
	fi
	if ! printf '%s\n' "$active_config" | grep -Eq '^[[:space:]]*read_header[[:space:]]+10s[[:space:]]*$'; then
		echo "$caddyfile must keep the 10-second request-header timeout" >&2
		exit 1
	fi
	if ! printf '%s\n' "$active_config" | grep -Eq '^[[:space:]]*idle[[:space:]]+2m[[:space:]]*$'; then
		echo "$caddyfile must keep the two-minute idle timeout" >&2
		exit 1
	fi
	if ! printf '%s\n' "$active_config" | grep -Eq '^[[:space:]]*max_size[[:space:]]+100MB[[:space:]]*$'; then
		echo "$caddyfile must keep NanaFox's 100 MB edge request-body limit" >&2
		exit 1
	fi
	if printf '%s\n' "$active_config" | grep -Eq '^[[:space:]]*flush_interval([[:space:]]|$)'; then
		echo "$caddyfile must leave flush_interval unset so SSE auto-flushing and client cancellation remain intact" >&2
		exit 1
	fi
	if printf '%s\n' "$active_config" | grep -Eq '^[[:space:]]*header[[:space:]]+Content-Type[[:space:]]+text/\*'; then
		echo "$caddyfile must not compress every text response because that buffers SSE" >&2
		exit 1
	fi
	if ! printf '%s\n' "$active_config" | grep -Eq '^[[:space:]]*header[[:space:]]+Content-Type[[:space:]]+text/plain\*'; then
		echo "$caddyfile must keep the explicit non-SSE text compression policy" >&2
		exit 1
	fi
done

echo "NanaFox Caddyfiles preserve ingress limits, backend cache policy, SSE streaming, and reverse_proxy routing"
