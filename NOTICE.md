# Third-party material

The FSD server sources in `fsd/` remain under the terms described in
[README.md](README.md). The notices below apply only to the WebEye frontend in
`webeye/`.

## VATSIM Radar: aircraft icons and icon mapping

`webeye/src/public/aircraft/*.svg` and
`webeye/src/src/assets/aircraft-icons.json` are taken from
[VATSIM Radar](https://github.com/VATSIM-Radar/vatsim-radar).

The colour palette in `webeye/src/src/style.css` is derived from the same
project's design tokens.

> Copyright © VATSIM Radar contributors
> Licensed under [Creative Commons Attribution-NonCommercial 4.0 International](https://creativecommons.org/licenses/by-nc/4.0/)

**This is a NonCommercial licence.** Redistributing or deploying WebEye with
these assets is permitted for non-commercial purposes only. If you intend to
use WebEye commercially, remove `webeye/src/public/aircraft/` and
`aircraft-icons.json` and substitute your own artwork. `webeye/src/src/aircraft.ts`
falls back cleanly when an icon is missing.

## VATGlasses: airspace sector data

`webeye/data/vatglasses/*.json` is taken from
[lennycolton/vatglasses-data](https://github.com/lennycolton/vatglasses-data),
the dataset behind [VATGlasses](https://vatglasses.uk/).

> Copyright © VATGlasses contributors
> Licensed under [Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International](https://creativecommons.org/licenses/by-nc-sa/4.0/)

**This licence is ShareAlike as well as NonCommercial**, which is stricter than
the VATSIM Radar assets above: any adapted version of this data must itself be
released under the same terms. Only `ed.json` (Germany) is bundled; add more
regions at runtime with `WEBEYE_VATGLASSES`, or delete the directory entirely.
WebEye falls back to drawing a plain label for radar positions with no sector
data, and every other feature keeps working.

## OurAirports: airport coordinates

`webeye/data/airports.csv` is trimmed from the
[OurAirports](https://ourairports.com/data/) database, which its authors have
released into the **public domain**. Columns are ICAO identifier, latitude and
longitude.

## Leaflet

Bundled via npm. Copyright © Volodymyr Agafonkin and CloudMade, BSD-2-Clause.

## Basemap tiles

Rendered at runtime by [CARTO](https://carto.com/attributions) from
[OpenStreetMap](https://www.openstreetmap.org/copyright) data. Tiles are
requested by the viewer's browser and are not redistributed in this repository.
