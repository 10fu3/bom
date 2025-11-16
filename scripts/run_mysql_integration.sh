#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE_NAME="${IMAGE_NAME:-bom-mysql}"
CONTAINER_NAME="${CONTAINER_NAME:-bom-mysql-test}"
HOST_PORT="${HOST_PORT:-3306}"
DDL="${ROOT_DIR}/examples/mysql/schema.sql"
CONFIG_MYSQL="${ROOT_DIR}/examples/mysql/bom.yml"
GOCACHE_DIR="${GOCACHE:-${ROOT_DIR}/.gocache}"

# Enable SQL debug logging unless explicitly disabled.
if [[ -z "${BOM_DEBUG_SQL:-}" ]]; then
  if [[ -n "${SQL_DEBUG:-}" ]]; then
    export BOM_DEBUG_SQL="${SQL_DEBUG}"
  else
    export BOM_DEBUG_SQL=1
  fi
fi
echo "BOM_DEBUG_SQL=${BOM_DEBUG_SQL} (SQL statements will be logged)"

docker build -t "${IMAGE_NAME}" \
  -f "${ROOT_DIR}/examples/mysql/docker/mysql/Dockerfile" \
  "${ROOT_DIR}/examples/mysql/docker/mysql"

cleanup() {
  docker rm -f "${CONTAINER_NAME}" >/dev/null 2>&1 || true
}
trap cleanup EXIT

GOCACHE="${GOCACHE_DIR}" go run ./cmd/bomgen --ddl "${DDL}" --config "${CONFIG_MYSQL}" --out "${ROOT_DIR}/examples/mysql/generated"

docker run -d --name "${CONTAINER_NAME}" -p "${HOST_PORT}:3306" --tmpfs /var/lib/mysql "${IMAGE_NAME}" >/dev/null

echo "Waiting for MySQL to accept connections..."
for _ in {1..30}; do
  if docker exec "${CONTAINER_NAME}" mysqladmin ping -uroot >/dev/null 2>&1; then
    break
  fi
  sleep 2
done

if ! docker exec "${CONTAINER_NAME}" mysqladmin ping -uroot >/dev/null 2>&1; then
  echo "MySQL did not become ready in time" >&2
  exit 1
fi

export TEST_MYSQL_DSN="root@tcp(127.0.0.1:${HOST_PORT})/bom_test?parseTime=true"
cd "${ROOT_DIR}"
GOCACHE="${GOCACHE_DIR}" go test -v -count=1 -tags mysqlserver ./examples/mysql
