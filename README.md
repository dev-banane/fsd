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
| `FSD_WEATHER_SOURCE` | `source` | `file` |
| `FSD_CERTS` | `cert.txt` | *(unset — keep existing file)* |
| `FSD_MOTD` | `motd.txt` | *(unset — keep existing file)* |
| `TZ` | — | `UTC` |

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

