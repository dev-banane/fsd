# FSD 4.0

## History

This repository is derived from the last known public copy of 
Marty Bochane's FSD 2 source code.

It originated from a zip file 'fsd1110.zip', which bears indications it may
have been "FSFDT Windows FSD Beta from FSD V3.000 draft 9".
It is almost identical to contents of the fsd-ubuntu-120413.tar.bz2 distribution
still available from the Apollo3 flight simulator site.

## License

No license was included with the source code, nor copyright claim.

The sources are assumed to be available under the public domain.

Unfortunately, as an Australian, any modifications I make on this code are
automatically copyrighted, and as such, I release my modifications under
the Free Public License.

Modifications can be identified by comparing the master branch against
`original-fsd-2`

### The Free Public License

Permission to use, copy, modify, and/or distribute this software for any 
purpose with or without fee is hereby granted.

THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH 
REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY 
AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT, 
INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM 
LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR 
OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR 
PERFORMANCE OF THIS SOFTWARE.

## Docker

FSD is a raw TCP service. Clients connect to **TCP 6809**.

```sh
docker compose up --build
```

Persistent data (`fsd.conf`, `cert.txt`, `motd.txt`, logs, METAR) is stored in
`/data`.

Runtime settings can be injected with environment variables:

| Variable | Config key | Default |
| --- | --- | --- |
| `FSD_IDENT` | `ident` | `FSD` |
| `FSD_HOSTNAME` | `hostname` | `localhost` |
| `FSD_PASSWORD` | `password` | `disable` |
| `FSD_EMAIL` | `email` | `nobody@nowhere.com` |
| `FSD_NAME` | `name` | `FSD Docker` |
| `FSD_LOCATION` | `location` | `Nowhere` |
| `FSD_MAXCLIENTS` | `maxclients` | `200` |
| `FSD_WEATHER_SOURCE` | `source` | `file` (empty `metar.txt`). Use `download` for live NOAA METAR |
| `FSD_WHAZZUP_INTERVAL` | `whazzupinterval` | `5` |
| `FSD_CERTS` | `cert.txt` | *(unset; keep existing file)* |
| `FSD_MOTD` | `motd.txt` | *(unset; keep existing file)* |
| `TZ` | n/a | `UTC` |

Change `FSD_PASSWORD` before exposing the system port. Set `FSD_HOSTNAME` to
the public hostname or IP that clients should use.

`FSD_CERTS` writes `/data/cert.txt` at startup. The native file format is
whitespace-separated `cid password level`, where `level` is a number (`12` =
administrator). The environment variable uses:

```text
FSD_CERTS=100000:MyPassword
FSD_CERTS=100000:password1,100001:password2
FSD_CERTS=100000:MyPassword:12
```

Level defaults to `12` when omitted. If `FSD_CERTS` is unset, an existing
`/data/cert.txt` is left unchanged. CID and password must not contain spaces,
commas, or colons.

`FSD_MOTD` writes `/data/motd.txt`, the welcome text EuroScope shows after
connect. Use `\n` for line breaks:

```text
FSD_MOTD=Welcome to my FSD server.\nEnjoy.
```

If `FSD_MOTD` is unset, an existing `/data/motd.txt` is left unchanged.

### Certificate levels

The number in `cert.txt` is the FSD/VATSIM **controller rating**. EuroScope
sends the rating from its Connect dialog; FSD rejects the login with
"Requested level too high" if that rating is above the certificate.

VATSIM still uses these IDs. `C2` and `I2` exist in FSD but are unused on
the live network. This is **not** the separate VATSIM pilot rating
(PPL/IR/ATPL).

| Level | FSD name | VATSIM | Typical meaning |
| --- | --- | --- | --- |
| 0 | SUSPENDED | SUS | Account suspended |
| 1 | OBSPILOT | OBS | Observe only |
| 2 | STUDENT1 | S1 | Delivery / ground |
| 3 | STUDENT2 | S2 | Tower |
| 4 | STUDENT3 | S3 | Approach / departure |
| 5 | CONTROLLER1 | C1 | Enroute / center |
| 6 | CONTROLLER2 | C2 | Unused on VATSIM |
| 7 | CONTROLLER3 | C3 | Senior controller |
| 8 | INSTRUCTOR1 | I1 | Instructor |
| 9 | INSTRUCTOR2 | I2 | Unused on VATSIM |
| 10 | INSTRUCTOR3 | I3 | Senior instructor |
| 11 | SUPERVISOR | SUP | Network supervisor |
| 12 | ADMINISTRATOR | ADM | Administrator |

For a private EuroScope server, `12` is the practical default: any rating
in the Connect dialog will be accepted.

The compose file publishes `6809/tcp`. Optionally also publish `3010/tcp` for
the telnet system console (`pwd <password>` then `help`).

## Protocol additions

Ported from [Tallerik/BetterFSD](https://github.com/Tallerik/BetterFSD):

- **Fast position updates.** `^` (`CL_FASTPOS`) is relayed as a broadcast and,
  unlike every other client command, carries no destination field. Pilot
  clients that support it send smoother position data than `@` alone gives.
- **Visual update handshake.** On pilot connect the server now sends
  `$SFSERVER:<callsign>:1` followed by `$CQSERVER:<callsign>:CAPS`, which is
  what tells a client to start sending `^`.
- **`pbh` in WhazzUp.** Each client row gains a trailing packed
  pitch/bank/heading word, so consumers can show aircraft heading.
- **METAR host** moved to `tgftp.nws.noaa.gov`; `weather.noaa.gov` is dead.

## WebEye: live map

`docker compose up --build` also starts **webeye**, a live traffic map on
<http://localhost:8080>.

WebEye never talks to FSD directly. FSD writes `/data/whazzup.txt` every
`whazzupinterval` seconds; WebEye reads that file from the same volume,
mounted read-only, and serves it as JSON plus a map.

- **Backend:** Go, standard library only. Parses WhazzUp, keeps position
  history, resolves airport coordinates, serves the API and the frontend. The
  built frontend is embedded with `go:embed`, so the binary is self-contained.
- **Frontend:** `webeye/src/`, Vite + Vue 3 + TypeScript + Tailwind v4,
  Leaflet for the map. Vite builds into `webeye/static/`, which the binary
  embeds at compile time.

**Controllers are the focus**, laid out the way VATSIM Radar does it:

- **Airport stations:** DEL, GND, TWR and ATIS are drawn *at their airport*,
  not where the controller reports being: an ICAO label with a row of coloured
  pills beneath it, one per staffed position. The airport is the first callsign
  token and the position is the last, so `EDDM_TWR`, `EDDM_N_TWR` and
  `EDDM_X_TWR` all land on the same EDDM station. Middle tokens are ignored.
- **Sectors:** APP and CTR are drawn as **real VATGlasses airspace polygons**,
  in that sector's own VATGlasses colour, with the callsign as plain coloured
  text and no label box.

Sector ownership follows VATGlasses' own rules: every airspace volume carries an
ordered list of positions, and the volume goes to the highest-priority position
that is online. So a centre picks up the sectors nobody is staffing, and hands
them back when that controller connects. A controller who logs in without a
sector id (`EDDM_APP` rather than `EDMM_ALB_CTR`) claims every matching position
sharing the prefix.

Radar positions with no VATGlasses match get a plain label where they sit, not
an invented circle. Pilots are secondary and can be toggled off the map.

**Sectors are sliced by altitude.** VATGlasses airspace is stacked in bands, so
drawing every volume at once is unreadable. The level selector in the centre of
the top bar picks a flight level and only the volumes owning that level are
drawn. A Munich approach sector disappears above roughly FL195, while the
centre sectors above it expand. Volumes the data leaves open-topped (`max: 0`)
are treated as unlimited rather than being filtered out.

| Route | Purpose |
| --- | --- |
| `/` | Map: controller sectors, aircraft with heading, trails, routes |
| `/fsd-status` | JSON snapshot of the current traffic |
| `/healthz` | `{"ok": true}` once a WhazzUp file has been read |

| Variable | Meaning | Default |
| --- | --- | --- |
| `WEBEYE_PORT` | Published port | `8080` |
| `WEBEYE_TITLE` | Map header and browser tab title | `WebEye` |
| `WEBEYE_POLL` | Seconds between file reads | `5` |
| `WEBEYE_HISTORY_POINTS` | Trail length per aircraft | `60` |
| `WEBEYE_HISTORY_TTL` | Seconds a disconnected trail is kept | `600` |
| `WEBEYE_VATGLASSES` | Directory of extra VATGlasses region files | *(bundled only)* |

Only the German dataset (`ed.json`) is bundled, because the full VATGlasses
database is ~29 MB. To cover more airspace, drop further region files into a
directory and point `WEBEYE_VATGLASSES` at it; they are merged over the
bundled set:

```bash
curl -O https://raw.githubusercontent.com/lennycolton/vatglasses-data/main/data/eg.json
```

Refresh rate is bounded by `FSD_WHAZZUP_INTERVAL`. WebEye cannot show data
faster than FSD writes it. Both default to 5 seconds.

Aircraft heading and the on-ground flag come from the `pbh` field, a packed
pitch/bank/heading word appended to each WhazzUp client row.

Selecting an aircraft draws its route as two great-circle legs: departure to
aircraft solid, aircraft to destination dashed. Airport coordinates come from a
trimmed OurAirports table (19,171 ICAO codes) embedded in the binary, so this
needs no external lookup; a route leg is simply omitted when the ICAO code is
unknown.

The basemap is CARTO's free vector style rendered with MapLibre GL inside
Leaflet, patched so every label prefers its English name (`webeye/src/src/basemap.ts`).
The unpatched style otherwise stacks several scripts on large water bodies
and countries. Browsers viewing the map need internet access for tiles,
glyphs and the style itself; if that fetch fails, WebEye falls back to plain
CARTO raster tiles (no English forcing, since raster has no per-label
language to select) rather than showing a blank map. Point `STYLE_URL` and
`FALLBACK_TILES` in `webeye/src/src/basemap.ts` / `TrafficMap.vue` at your own
tile server to run fully offline.

To run the server without the map, start just the one service:

```bash
docker compose up --build fsd
```

### Development

Build the frontend first. The Go binary embeds `webeye/static`, so it will
not compile without it:

```bash
cd webeye/src && npm install && npm run build
```

Then build and run the backend against a WhazzUp file:

```bash
cd webeye && go build -o webeye . && WEBEYE_WHAZZUP=/path/to/whazzup.txt ./webeye
```

For frontend work, run Vite instead; it proxies `/fsd-status` to the running
Go server and gives you hot reload. Set `WEBEYE_API` if the backend is not on
`http://127.0.0.1:8080`:

```bash
cd webeye/src && npm run dev
```

Go tests:

```bash
cd webeye && go test ./...
```

### Attribution and licensing

Two bundled datasets restrict how WebEye may be used:

| Source | Used for | Licence |
| --- | --- | --- |
| [VATSIM Radar](https://github.com/VATSIM-Radar/vatsim-radar) | Aircraft icons, colour palette | CC BY-NC 4.0 |
| [VATGlasses](https://github.com/lennycolton/vatglasses-data) | Sector polygons | CC BY-NC-**SA** 4.0 |
| [OurAirports](https://ourairports.com/data/) | Airport coordinates | Public domain |

Both CC licences are **non-commercial**, and the VATGlasses one adds
**ShareAlike**. See [NOTICE.md](NOTICE.md) for the full terms and for how to
strip either dataset. WebEye degrades gracefully without them.

