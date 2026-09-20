#!/usr/bin/env bash
# Check the TLS certificate Caddy serves for the local domain, and verify the
# chain against Caddy's own local CA (not the host trust store).
set -euo pipefail

WAIT=false
if [[ "${1:-}" == "--wait" ]]; then
	WAIT=true
	shift
fi

HOST="${1:-matchmaker.localhost}"
PORT="${2:-443}"
CADDY="${CADDY:-caddy}"

# fetch_leaf_info prints dates/subject/issuer for the cert currently served.
fetch_leaf_info() {
	echo | openssl s_client -connect "127.0.0.1:${PORT}" -servername "$HOST" 2>/dev/null \
		| openssl x509 -noout -dates -subject -issuer -checkend 0 2>&1
}

if $WAIT; then
	echo "Waiting for valid TLS cert on https://${HOST}..."
	cert_info=""
	for _ in $(seq 1 30); do
		if cert_info="$(fetch_leaf_info 2>/dev/null)" && [[ -n "$cert_info" ]]; then
			break
		fi
		sleep 1
	done
	if [[ -z "$cert_info" ]]; then
		echo "Timed out waiting for a valid TLS certificate" >&2
		exit 1
	fi
else
	cert_info="$(fetch_leaf_info 2>&1 || true)"
fi

echo "TLS check for https://${HOST}"
echo "$cert_info"

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

# Verify against Caddy's local CA. Host openssl has no reason to trust that
# CA, so a system-store check reports "unable to get local issuer certificate"
# even when the signature and chain are fine.
if docker exec "$CADDY" cat /data/caddy/pki/authorities/local/root.crt >"$tmpdir/root.crt" 2>/dev/null \
	&& docker exec "$CADDY" cat /data/caddy/pki/authorities/local/intermediate.crt >"$tmpdir/int.crt" 2>/dev/null \
	&& echo | openssl s_client -connect "127.0.0.1:${PORT}" -servername "$HOST" 2>/dev/null \
		| openssl x509 >"$tmpdir/leaf.crt" 2>/dev/null; then
	root_fp="$(openssl x509 -in "$tmpdir/root.crt" -noout -fingerprint -sha256 | sed 's/^.*Fingerprint=//')"
	echo "Caddy root SHA256: $root_fp"
	if openssl verify -CAfile "$tmpdir/root.crt" -untrusted "$tmpdir/int.crt" "$tmpdir/leaf.crt" >/dev/null 2>&1; then
		echo "Verify: 0 (ok against Caddy local CA)"
		echo
		echo "If Firefox shows SEC_ERROR_BAD_SIGNATURE, the site cert is valid — Firefox"
		echo "is using a stale Caddy CA with the same name. Delete every \"Caddy Local"
		echo "Authority\" entry under Settings → Certificates → Authorities, then:"
		echo "  make export-ca"
		echo "and re-import caddy-root.crt (trust for websites). Restart Firefox."
	else
		echo "Verify: failed against Caddy local CA" >&2
		openssl verify -CAfile "$tmpdir/root.crt" -untrusted "$tmpdir/int.crt" "$tmpdir/leaf.crt" >&2 || true
		exit 1
	fi
else
	verify_code="$(echo | openssl s_client -connect "127.0.0.1:${PORT}" -servername "$HOST" 2>/dev/null \
		| grep -oP 'Verify return code: \K[0-9]+ \([^)]+\)' | head -1 || echo "unknown")"
	echo "Verify (system trust store): $verify_code"
	echo "Could not load Caddy local CA from container '${CADDY}'." >&2
fi
