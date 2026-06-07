#!/usr/bin/env bash
# Check TLS certificate served by Caddy for matchmaker.localhost.
set -euo pipefail

WAIT=false
if [[ "${1:-}" == "--wait" ]]; then
	WAIT=true
	shift
fi

HOST="${1:-matchmaker.localhost}"
PORT="${2:-443}"

fetch_cert() {
	echo | openssl s_client -connect "127.0.0.1:${PORT}" -servername "$HOST" 2>/dev/null \
		| openssl x509 -noout -dates -subject -issuer -checkend 0 2>&1
}

if $WAIT; then
	echo "Waiting for valid TLS cert on https://${HOST}..."
	cert_info=""
	for _ in $(seq 1 30); do
		if cert_info="$(fetch_cert 2>/dev/null)" && [[ -n "$cert_info" ]]; then
			break
		fi
		sleep 1
	done
	if [[ -z "$cert_info" ]]; then
		echo "Timed out waiting for a valid TLS certificate" >&2
		exit 1
	fi
else
	cert_info="$(fetch_cert 2>&1 || true)"
fi

verify_code="$(echo | openssl s_client -connect "127.0.0.1:${PORT}" -servername "$HOST" 2>/dev/null \
	| grep -oP 'Verify return code: \K[0-9]+ \([^)]+\)' | head -1 || echo "unknown")"

echo "TLS check for https://${HOST}"
echo "$cert_info"
echo "Verify: $verify_code"
