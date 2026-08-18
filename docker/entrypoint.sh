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
copy_if_missing "${TEMPLATE_DIR}/help.txt" "${DATA_DIR}/help.txt"

if [ -n "${FSD_MOTD:-}" ]; then
  printf '%b\n' "$FSD_MOTD" > "${DATA_DIR}/motd.txt"
else
  copy_if_missing "${TEMPLATE_DIR}/motd.txt" "${DATA_DIR}/motd.txt"
fi

write_certs() {
  dest=$1
  certs=$2
  tmp="${dest}.tmp"
  : > "$tmp"
  rest=$certs
  while [ -n "$rest" ]; do
    case "$rest" in
      *,*)
        entry=${rest%%,*}
        rest=${rest#*,}
        ;;
      *)
        entry=$rest
        rest=
        ;;
    esac
    [ -n "$entry" ] || continue
    cid=${entry%%:*}
    remainder=${entry#*:}
    if [ "$cid" = "$entry" ] || [ -z "$remainder" ]; then
      echo "FSD_CERTS entry '$entry' must be cid:password or cid:password:level" >&2
      exit 1
    fi
    case "$remainder" in
      *:*)
        password=${remainder%%:*}
        level=${remainder#*:}
        ;;
      *)
        password=$remainder
        level=12
        ;;
    esac
    if [ -z "$cid" ] || [ -z "$password" ] || [ -z "$level" ]; then
      echo "FSD_CERTS entry '$entry' must be cid:password or cid:password:level" >&2
      exit 1
    fi
    printf '%s %s %s\n' "$cid" "$password" "$level" >> "$tmp"
  done
  mv "$tmp" "$dest"
}

if [ -n "${FSD_CERTS:-}" ]; then
  write_certs "${DATA_DIR}/cert.txt" "$FSD_CERTS"
else
  [ -f "${DATA_DIR}/cert.txt" ] || : > "${DATA_DIR}/cert.txt"
fi

set_conf() {
  key=$1
  value=$2
  file=$3
  section=${4:-[system]}
  if grep -q "^${key}=" "$file"; then
    awk -v k="$key" -v v="$value" '
      index($0, k "=") == 1 { print k "=" v; next }
      { print }
    ' "$file" > "${file}.tmp"
  else
    awk -v k="$key" -v v="$value" -v sec="$section" '
      { print }
      $0 == sec { print k "=" v }
    ' "$file" > "${file}.tmp"
  fi
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
[ -n "${FSD_WHAZZUP_INTERVAL:-}" ] && set_conf whazzupinterval "$FSD_WHAZZUP_INTERVAL" "$CONFIG"

METAR_URL="${FSD_METAR_URL:-https://metar.vatsim.net/all}"
METAR_INTERVAL="${FSD_METAR_INTERVAL:-600}"
METAR_FILE="${DATA_DIR}/metar.txt"
fetch_metar=0

case "${FSD_WEATHER_SOURCE:-}" in
  download)
    fetch_metar=1
    set_conf source file "$CONFIG" "[weather]"
    ;;
  ftp)
    set_conf source download "$CONFIG" "[weather]"
    ;;
  ?*)
    set_conf source "$FSD_WEATHER_SOURCE" "$CONFIG" "[weather]"
    ;;
esac

fetch_metar_once() {
  tmp="${METAR_FILE}.tmp"
  if curl -fsS --max-time 60 "$METAR_URL" -o "$tmp" && [ -s "$tmp" ]; then
    mv "$tmp" "$METAR_FILE"
    chown fsd:fsd "$METAR_FILE"
    return 0
  fi
  rm -f "$tmp"
  return 1
}

if [ "$fetch_metar" = 1 ]; then
  echo "METAR: fetching $METAR_URL every ${METAR_INTERVAL}s" >&2
  fetch_metar_once || echo "METAR: initial fetch from $METAR_URL failed" >&2
  (
    while :; do
      sleep "$METAR_INTERVAL"
      fetch_metar_once || echo "METAR: fetch from $METAR_URL failed" >&2
    done
  ) &
fi

chown -R fsd:fsd "$DATA_DIR"

cd "$DATA_DIR"

echo "Starting FSD (client TCP ${FSD_CLIENTPORT:-6809})" >&2
exec gosu fsd /usr/local/bin/fsd -d -f "$CONFIG"
