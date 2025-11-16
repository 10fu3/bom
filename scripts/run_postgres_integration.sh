#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE_NAME="${PG_IMAGE_NAME:-bom-postgres}"
CONTAINER_NAME="${PG_CONTAINER_NAME:-bom-postgres-test}"
HOST_PORT="${PG_HOST_PORT:-5432}"
DDL="${ROOT_DIR}/examples/postgres/schema.sql"
CONFIG_POSTGRES="${ROOT_DIR}/examples/postgres/bom.yml"
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
  -f "${ROOT_DIR}/examples/postgres/docker/postgres/Dockerfile" \
  "${ROOT_DIR}/examples/postgres/docker/postgres"

cleanup() {
  docker rm -f "${CONTAINER_NAME}" >/dev/null 2>&1 || true
}
trap cleanup EXIT

GOCACHE="${GOCACHE_DIR}" go run ./cmd/bomgen --ddl "${DDL}" --config "${CONFIG_POSTGRES}" --parser mysql --out "${ROOT_DIR}/examples/postgres/generated"

docker run -d --name "${CONTAINER_NAME}" -p "${HOST_PORT}:5432" --tmpfs /var/lib/postgresql/data "${IMAGE_NAME}" >/dev/null

echo "Waiting for Postgres to accept connections..."
for _ in {1..30}; do
  if docker exec "${CONTAINER_NAME}" pg_isready -U postgres >/dev/null 2>&1; then
    break
  fi
  sleep 2
done

if ! docker exec "${CONTAINER_NAME}" pg_isready -U postgres >/dev/null 2>&1; then
  echo "Postgres did not become ready in time" >&2
  exit 1
fi

export TEST_POSTGRES_DSN="postgres://postgres@127.0.0.1:${HOST_PORT}/bom_test?sslmode=disable"
cd "${ROOT_DIR}"
GOCACHE="${GOCACHE_DIR}" go test -v -count=1 -tags postgresserver ./examples/postgres
