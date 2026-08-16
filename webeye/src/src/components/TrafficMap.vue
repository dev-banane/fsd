<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from "vue";
import L from "leaflet";
import { maplibreGL } from "@maplibre/maplibre-gl-leaflet";
import { positionColor, type Controller, type Pilot, type Position, type Status } from "../api";
import { aircraftMarkup, aircraftSize } from "../aircraft";
import { greatCircle, type Point } from "../geo";
import { loadEnglishBasemapStyle } from "../basemap";
import {
  buildStations,
  isSectorPosition,
  positionLetter,
  type Station,
} from "../stations";

const props = defineProps<{
  status: Status;
  selected: string | null;
  showPilots: boolean;
  showSectors: boolean;
  level: number;
}>();
const emit = defineEmits<{ select: [callsign: string | null] }>();

const el = ref<HTMLDivElement | null>(null);
let map: L.Map | undefined;
let resizeObserver: ResizeObserver | undefined;
let zoomControl: L.Control | undefined;
let zoomMqCleanup: (() => void) | undefined;
const planes = new Map<string, L.Marker>();
const trails = new Map<string, L.Polyline>();
const sectorShapes = new Map<string, L.Polygon>();
const sectorLabels = new Map<string, L.Marker>();
const stationMarkers = new Map<string, L.Marker>();
let routeLayers: L.Polyline[] = [];

const ATTRIBUTION = '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> &copy; <a href="https://carto.com/attributions">CARTO</a>';
const FALLBACK_TILES = "https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png";
const DESKTOP_MQ = "(min-width: 768px)";

const placed = (o: { lat: number | null; lon: number | null }) =>
  o.lat != null && o.lon != null && !(o.lat === 0 && o.lon === 0);

const popupOpts = (): L.PopupOptions => {
  const mobile = typeof window !== "undefined" && !window.matchMedia(DESKTOP_MQ).matches;
  return {
    maxWidth: 300,
    autoPanPaddingTopLeft: L.point(16, mobile ? 12 : 20),
    autoPanPaddingBottomRight: L.point(16, mobile ? 88 : 20),
  };
};

const esc = (s: unknown) =>
  String(s ?? "").replace(/[&<>"]/g, (c) =>
    ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;" })[c] as string,
  );

function card(title: string, body: string, extra?: { sub?: string; banner?: string }) {
  return `
    <div class="we-card">
      <div class="we-card-head">
        <div class="we-card-title">${esc(title)}</div>
        ${extra?.sub ? `<div class="we-card-sub">${esc(extra.sub)}</div>` : ""}
      </div>
      ${extra?.banner ? `<div class="we-card-route">${esc(extra.banner)}</div>` : ""}
      <div class="we-card-body">${body}</div>
    </div>`;
}

function kvBody(rows: [string, string][]) {
  return rows
    .filter(([, v]) => v)
    .map(
      ([k, v]) =>
        `<div class="we-card-row">
           <span class="we-card-k">${esc(k)}</span>
           <span class="we-card-v nums">${esc(v)}</span>
         </div>`,
    )
    .join("");
}

function facilityChip(position: Position) {
  return `<span class="we-facility" style="background:${positionColor[position]}">${esc(position)}</span>`;
}

function controllerRows(controllers: Controller[]) {
  return controllers
    .map(
      (c) =>
        `<div class="we-atc-row">
           <div class="we-atc-main">
             ${facilityChip(c.position)}
             <span class="we-atc-callsign">${esc(c.callsign)}</span>
             <span class="we-atc-freq nums">${esc(c.frequency || "")}</span>
           </div>
           ${c.name ? `<div class="we-atc-name">${esc(c.name)}</div>` : ""}
         </div>`,
    )
    .join("");
}

function stationIcon(station: Station, active: boolean): L.DivIcon {
  const pills = station.positions
    .map((p) => {
      const on = active && station.controllers.some((c) => c.position === p && c.callsign === props.selected);
      return `<span class="we-station-pill${on ? " we-station-pill--active" : ""}"
                    style="background:${positionColor[p]}">${positionLetter[p]}</span>`;
    })
    .join("");

  return L.divIcon({
    className: "",
    iconSize: [0, 0],
    iconAnchor: [0, 0],
    html: `
      <div class="we-station${active ? " we-station--active" : ""}">
        <span class="we-station-icao">${esc(station.icao)}</span>
        <span class="we-station-pills">${pills}</span>
      </div>`,
  });
}

function stationPopup(station: Station) {
  return card(station.icao, controllerRows(station.controllers), {
    sub: station.controllers.length === 1 ? "1 position" : `${station.controllers.length} positions`,
  });
}

function renderStations() {
  if (!map) return;
  const { stations, unplaced } = buildStations(props.status.controllers);
  const live = new Set<string>();

  for (const station of stations) {
    live.add(station.icao);
    const active = station.controllers.some((c) => c.callsign === props.selected);
    const centre: L.LatLngExpression = [station.pos.lat, station.pos.lon];

    const existing = stationMarkers.get(station.icao);
    if (existing) {
      existing.setLatLng(centre);
      existing.setIcon(stationIcon(station, active));
      existing.setPopupContent(stationPopup(station));
    } else {
      stationMarkers.set(
        station.icao,
        L.marker(centre, {
          icon: stationIcon(station, active),
          pane: "atcLabels",
          riseOnHover: true,
        })
          .bindPopup(stationPopup(station), popupOpts())
          .on("click", () => emit("select", station.controllers[0].callsign))
          .addTo(map!),
      );
    }
  }

  for (const c of unplaced.filter(placed)) {
    const key = `~${c.callsign}`;
    live.add(key);
    const pseudo: Station = {
      icao: c.airport || c.callsign,
      pos: { lat: c.lat!, lon: c.lon! },
      controllers: [c],
      positions: [c.position],
    };
    const active = props.selected === c.callsign;
    const existing = stationMarkers.get(key);
    if (existing) {
      existing.setLatLng([c.lat!, c.lon!]);
      existing.setIcon(stationIcon(pseudo, active));
      existing.setPopupContent(stationPopup(pseudo));
    } else {
      stationMarkers.set(
        key,
        L.marker([c.lat!, c.lon!], {
          icon: stationIcon(pseudo, active),
          pane: "atcLabels",
          riseOnHover: true,
        })
          .bindPopup(stationPopup(pseudo), popupOpts())
          .on("click", () => emit("select", c.callsign))
          .addTo(map!),
      );
    }
  }

  for (const [key, marker] of stationMarkers)
    if (!live.has(key)) (map.removeLayer(marker), stationMarkers.delete(key));
}

function sectorLabelIcon(c: Controller, active: boolean, vgColor?: string): L.DivIcon {
  const color = sectorColor(c, vgColor);
  return L.divIcon({
    className: "we-sector-marker",
    iconSize: [0, 0],
    iconAnchor: [0, 0],
    html: `
      <div class="we-sector-label${active ? " we-sector-label--active" : ""}"
           style="--we-sector:${color}">
        ${esc(c.callsign)}
      </div>`,
  });
}

const sectorPopup = (c: Controller, sectorName: string) =>
  card(c.callsign, kvBody([
    ["Controller", c.name],
    ["Position", c.position],
    ["Sector", sectorName],
  ]), { sub: c.frequency });

function sectorColor(c: Controller, vgColor?: string): string {
  return vgColor || positionColor[c.position];
}

function labelAnchor(polygons: [number, number][][]): L.LatLngExpression | null {
  let best: [number, number][] | null = null;
  let bestArea = -1;
  for (const poly of polygons) {
    const lats = poly.map((p) => p[0]);
    const lons = poly.map((p) => p[1]);
    const area =
      (Math.max(...lats) - Math.min(...lats)) * (Math.max(...lons) - Math.min(...lons));
    if (area > bestArea) (bestArea = area), (best = poly);
  }
  if (!best) return null;
  const lat = best.reduce((a, p) => a + p[0], 0) / best.length;
  const lon = best.reduce((a, p) => a + p[1], 0) / best.length;
  return [lat, lon];
}

function renderSectors() {
  if (!map) return;
  const live = new Set<string>();
  const visible = props.showSectors
    ? props.status.sectors.filter((s) => props.level >= s.min && props.level <= s.max)
    : [];

  // Group the volumes the backend resolved by the controller holding them.
  const byCallsign = new Map<string, { polygons: [number, number][][]; color: string; name: string }>();
  for (const s of visible) {
    let entry = byCallsign.get(s.callsign);
    if (!entry) {
      entry = { polygons: [], color: s.color, name: s.name };
      byCallsign.set(s.callsign, entry);
    }
    entry.polygons.push(s.polygon);
  }

  const controllers = new Map(props.status.controllers.map((c) => [c.callsign, c]));

  const spanOf = (polygons: [number, number][][]) => {
    let lo = 90, hi = -90, west = 180, east = -180;
    for (const poly of polygons)
      for (const [lat, lon] of poly) {
        lo = Math.min(lo, lat); hi = Math.max(hi, lat);
        west = Math.min(west, lon); east = Math.max(east, lon);
      }
    return (hi - lo) * (east - west);
  };
  const ordered = [...byCallsign].sort(
    ([, a], [, b]) => spanOf(b.polygons) - spanOf(a.polygons),
  );

  for (const [callsign, entry] of ordered) {
    const c = controllers.get(callsign);
    if (!c) continue;
    live.add(callsign);

    const active = props.selected === callsign;
    const color = sectorColor(c, entry.color);
    const style: L.PathOptions = {
      color,
      weight: active ? 2 : 1,
      opacity: active ? 1 : 0.6,
      fillColor: color,
      fillOpacity: 0.2,
      lineJoin: "round",
    };

    const multi = entry.polygons.map((ring) => [ring]);
    const shape = sectorShapes.get(callsign);
    if (shape) {
      shape.setLatLngs(multi).setStyle(style);
    } else {
      sectorShapes.set(
        callsign,
        L.polygon(multi, { ...style, pane: "atcSectors" })
          .on("click", () => emit("select", callsign))
          .addTo(map!),
      );
    }

    const anchor = labelAnchor(entry.polygons);
    if (!anchor) continue;
    const label = sectorLabels.get(callsign);
    if (label) {
      label.setLatLng(anchor).setIcon(sectorLabelIcon(c, active, entry.color));
    } else {
      sectorLabels.set(
        callsign,
        L.marker(anchor, {
          icon: sectorLabelIcon(c, active, entry.color),
          pane: "atcLabels",
        })
          .bindPopup(sectorPopup(c, entry.name), popupOpts())
          .on("click", () => emit("select", callsign))
          .addTo(map!),
      );
    }
  }

  const hasSectorData = new Set(props.status.sectors.map((s) => s.callsign));
  for (const c of props.showSectors ? props.status.controllers : []) {
    if (!isSectorPosition(c) || hasSectorData.has(c.callsign) || !placed(c)) continue;
    live.add(c.callsign);
    const active = props.selected === c.callsign;
    const label = sectorLabels.get(c.callsign);
    if (label) {
      label.setLatLng([c.lat!, c.lon!]).setIcon(sectorLabelIcon(c, active));
    } else {
      sectorLabels.set(
        c.callsign,
        L.marker([c.lat!, c.lon!], {
          icon: sectorLabelIcon(c, active),
          pane: "atcLabels",
        })
          .bindPopup(sectorPopup(c, "no VATGlasses data"), popupOpts())
          .on("click", () => emit("select", c.callsign))
          .addTo(map!),
      );
    }
  }

  for (const [cs, layer] of sectorShapes)
    if (!live.has(cs)) (map.removeLayer(layer), sectorShapes.delete(cs));
  for (const [cs, layer] of sectorLabels)
    if (!live.has(cs)) (map.removeLayer(layer), sectorLabels.delete(cs));
}

const pilotPopup = (p: Pilot) =>
  card(
    p.callsign,
    kvBody([
      ["Pilot", p.name],
      ["Aircraft", p.type],
      ["Altitude", p.alt ? `${p.alt.toLocaleString("en-US")} ft` : ""],
      ["Speed", p.groundspeed ? `${p.groundspeed} kt` : ""],
      ["Heading", `${Math.round(p.heading)}°`],
      ["Squawk", p.sqwk],
    ]),
    {
      banner: [p.departure_ap, p.arrival_ap].filter(Boolean).join("  →  ") || undefined,
    },
  );

async function renderPilots() {
  if (!map) return;

  if (!props.showPilots) {
    for (const [cs, m] of planes) (map.removeLayer(m), planes.delete(cs));
    for (const [cs, l] of trails) (map.removeLayer(l), trails.delete(cs));
    return;
  }

  const size = aircraftSize(map.getZoom());
  const live = new Set<string>();

  for (const p of props.status.pilots.filter(placed)) {
    live.add(p.callsign);
    const active = props.selected === p.callsign;
    const icon = L.divIcon({
      className: active ? "we-ac we-ac--active" : "we-ac",
      iconSize: [size, size],
      iconAnchor: [size / 2, size / 2],
      html: await aircraftMarkup(p.type, p.heading, "var(--ac)", size),
    });

    const existing = planes.get(p.callsign);
    if (existing) {
      existing.setLatLng([p.lat!, p.lon!]);
      existing.setIcon(icon);
      existing.setPopupContent(pilotPopup(p));
    } else {
      planes.set(
        p.callsign,
        L.marker([p.lat!, p.lon!], { icon, riseOnHover: true })
          .bindPopup(pilotPopup(p), popupOpts())
          .on("click", () => emit("select", p.callsign))
          .addTo(map!),
      );
    }
  }
  for (const [cs, m] of planes)
    if (!live.has(cs)) (map.removeLayer(m), planes.delete(cs));

  for (const [cs, points] of Object.entries(props.status.history ?? {})) {
    if (!live.has(cs) || points.length < 2) continue;
    const active = props.selected === cs;
    const style = {
      color: active ? "var(--color-ac-active)" : "var(--color-brand-stroke)",
      weight: active ? 2 : 1,
      opacity: active ? 0.85 : 0.3,
    };
    const line = trails.get(cs);
    if (line) line.setLatLngs(points as Point[]).setStyle(style);
    else
      trails.set(
        cs,
        L.polyline(points as Point[], { ...style, interactive: false }).addTo(map!),
      );
  }
  for (const [cs, line] of trails)
    if (!live.has(cs)) (map.removeLayer(line), trails.delete(cs));
}

function renderRoute() {
  if (!map) return;
  routeLayers.forEach((l) => map!.removeLayer(l));
  routeLayers = [];

  const p = props.status.pilots.find((x) => x.callsign === props.selected);
  if (!p || !placed(p)) return;

  const here: Point = [p.lat!, p.lon!];
  const legs: [Point, Point, boolean][] = [];
  if (p.departure_pos) legs.push([[p.departure_pos.lat, p.departure_pos.lon], here, false]);
  if (p.arrival_pos) legs.push([here, [p.arrival_pos.lat, p.arrival_pos.lon], true]);

  for (const [from, to, remaining] of legs) {
    routeLayers.push(
      L.polyline(greatCircle(from, to), {
        color: "var(--color-brand)",
        weight: 1.5,
        opacity: remaining ? 0.75 : 0.35,
        dashArray: remaining ? "6 6" : undefined,
        interactive: false,
      }).addTo(map),
    );
  }
}

function renderAll() {
  renderSectors();
  renderStations();
  renderRoute();
  void renderPilots();
}

function focus(callsign: string | null) {
  if (!map || !callsign) return;
  const target =
    planes.get(callsign) ??
    sectorLabels.get(callsign) ??
    [...stationMarkers.values()].find((m) => m.getPopup()?.getContent()?.toString().includes(callsign));
  if (target) map.panTo(target.getLatLng(), { animate: true, duration: 0.5 });
}

onMounted(() => {
  map = L.map(el.value!, {
    zoomControl: false,
    worldCopyJump: true,
    zoomSnap: 0.5,
    tapTolerance: 25,
    bounceAtZoomLimits: false,
    maxBounds: [
      [-90, -Infinity],
      [90, Infinity],
    ],
    maxBoundsViscosity: 1,
    minZoom: 1,
  }).setView([50, 9], 5);

  map.createPane("atcSectors").style.zIndex = "350";
  map.createPane("atcLabels").style.zIndex = "650";
  map.attributionControl.setPrefix(false);

  const zoomMq = window.matchMedia(DESKTOP_MQ);
  const placeZoom = () => {
    if (!map) return;
    if (zoomControl) map.removeControl(zoomControl);
    zoomControl = L.control
      .zoom({ position: zoomMq.matches ? "topleft" : "topright" })
      .addTo(map);
  };
  placeZoom();
  zoomMq.addEventListener("change", placeZoom);
  zoomMqCleanup = () => zoomMq.removeEventListener("change", placeZoom);

  loadEnglishBasemapStyle()
    .then((style) => {
      maplibreGL({ style }).addTo(map!);
    })
    .catch((err) => {
      console.warn("webeye: vector basemap unavailable, using raster fallback", err);
      L.tileLayer(FALLBACK_TILES, {
        subdomains: "abcd",
        maxZoom: 19,
        attribution: ATTRIBUTION,
      }).addTo(map!);
    });

  map.on("zoomend", () => {
    void renderPilots();
    renderStations();
    renderSectors();
  });

  resizeObserver = new ResizeObserver(() => map?.invalidateSize());
  resizeObserver.observe(el.value!);

  renderAll();
});

onUnmounted(() => {
  zoomMqCleanup?.();
  resizeObserver?.disconnect();
  map?.remove();
});

watch(() => props.status, renderAll);
watch(() => props.showPilots, renderAll);
watch(() => props.showSectors, renderAll);
watch(() => props.level, renderSectors);
watch(
  () => props.selected,
  (cs) => {
    renderAll();
    focus(cs);
  },
);
</script>

<template>
  <div ref="el" class="h-full w-full"></div>
</template>
