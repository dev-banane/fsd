#!/bin/sh
set -eu

DATA_DIR="${FSD_HOME:-/data}"
TEMPLATE_DIR="/opt/fsd"
CONFIG="${DATA_DIR}/fsd.conf"

mkdir -p "$DATA_DIR"

copy_if_missing() {
  if [ ! -f "$2" ]; then
    cp "$1" "$2"
  fi
}

copy_if_missing "${TEMPLATE_DIR}/fsd.conf" "$CONFIG"
copy_if_missing "${TEMPLATE_DIR}/motd.txt" "${DATA_DIR}/motd.txt"
copy_if_missing "${TEMPLATE_DIR}/help.txt" "${DATA_DIR}/help.txt"
[ -f "${DATA_DIR}/cert.txt" ] || : > "${DATA_DIR}/cert.txt"

set_conf() {
  key=$1
  value=$2
  file=$3
  awk -v k="$key" -v v="$value" '
    index($0, k "=") == 1 { print k "=" v; next }
    { print }
  ' "$file" > "${file}.tmp"
  mv "${file}.tmp" "$file"
}

[ -n "${FSD_CLIENTPORT:-}" ] && set_conf clientport "$FSD_CLIENTPORT" "$CONFIG"
[ -n "${FSD_SERVERPORT:-}" ] && set_conf serverport "$FSD_SERVERPORT" "$CONFIG"
[ -n "${FSD_SYSTEMPORT:-}" ] && set_conf systemport "$FSD_SYSTEMPORT" "$CONFIG"
[ -n "${FSD_IDENT:-}" ] && set_conf ident "$FSD_IDENT" "$CONFIG"
[ -n "${FSD_EMAIL:-}" ] && set_conf email "$FSD_EMAIL" "$CONFIG"
[ -n "${FSD_NAME:-}" ] && set_conf name "$FSD_NAME" "$CONFIG"
[ -n "${FSD_HOSTNAME:-}" ] && set_conf hostname "$FSD_HOSTNAME" "$CONFIG"
[ -n "${FSD_PASSWORD:-}" ] && set_conf password "$FSD_PASSWORD" "$CONFIG"
[ -n "${FSD_LOCATION:-}" ] && set_conf location "$FSD_LOCATION" "$CONFIG"
[ -n "${FSD_MAXCLIENTS:-}" ] && set_conf maxclients "$FSD_MAXCLIENTS" "$CONFIG"
[ -n "${FSD_WEATHER_SOURCE:-}" ] && set_conf source "$FSD_WEATHER_SOURCE" "$CONFIG"

chown -R fsd:fsd "$DATA_DIR"

cd "$DATA_DIR"

echo "Starting FSD (client TCP ${FSD_CLIENTPORT:-6809})" >&2
exec gosu fsd /usr/local/bin/fsd -d -f "$CONFIG"
