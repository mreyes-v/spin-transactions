#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

if [[ ! -f .env ]]; then
	cp .env.example .env
	echo "Creé .env desde .env.example"
fi

docker compose up -d --wait

write_pid=""
read_pid=""

cleanup() {
	local status=$?
	trap - EXIT INT TERM
	if [[ -n "$write_pid" ]]; then
		kill "$write_pid" 2>/dev/null || true
		wait "$write_pid" 2>/dev/null || true
	fi
	if [[ -n "$read_pid" ]]; then
		kill "$read_pid" 2>/dev/null || true
		wait "$read_pid" 2>/dev/null || true
	fi
	exit "$status"
}
trap cleanup EXIT INT TERM

echo "write  http://localhost:8080"
echo "read   http://localhost:8081"
echo "Ctrl-C detiene los dos procesos. Postgres sigue arriba; make db-down lo apaga y make db-reset lo vacía."

go run ./cmd/write &
write_pid=$!
HTTP_ADDR=:8081 go run ./cmd/read &
read_pid=$!

while true; do
	if ! kill -0 "$write_pid" 2>/dev/null; then
		wait "$write_pid" || exit $?
		exit 0
	fi
	if ! kill -0 "$read_pid" 2>/dev/null; then
		wait "$read_pid" || exit $?
		exit 0
	fi
	sleep 1
done
