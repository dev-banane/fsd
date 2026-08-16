import type { Controller, LatLon, Position } from "./api";

export const AIRPORT_POSITIONS: Position[] = ["DEL", "GND", "TWR", "ATIS"];
export const SECTOR_POSITIONS: Position[] = ["APP", "CTR"];

export const positionLetter: Record<Position, string> = {
  DEL: "D",
  GND: "G",
  TWR: "T",
  ATIS: "A",
  APP: "A",
  CTR: "C",
  OBS: "O",
};

export const PILL_ORDER: Position[] = ["DEL", "GND", "TWR", "ATIS"];

export interface Station {
  icao: string;
  pos: LatLon;
  controllers: Controller[];
  positions: Position[];
}

const isAirportPosition = (c: Controller) => AIRPORT_POSITIONS.includes(c.position);
export const isSectorPosition = (c: Controller) => SECTOR_POSITIONS.includes(c.position);
export function buildStations(controllers: Controller[]): {
  stations: Station[];
  unplaced: Controller[];
} {
  const byIcao = new Map<string, Station>();
  const unplaced: Controller[] = [];

  for (const c of controllers.filter(isAirportPosition)) {
    if (!c.airport || !c.airport_pos) {
      unplaced.push(c);
      continue;
    }
    let station = byIcao.get(c.airport);
    if (!station) {
      station = { icao: c.airport, pos: c.airport_pos, controllers: [], positions: [] };
      byIcao.set(c.airport, station);
    }
    station.controllers.push(c);
  }

  for (const station of byIcao.values()) {
    station.positions = PILL_ORDER.filter((p) =>
      station.controllers.some((c) => c.position === p),
    );
  }

  return {
    stations: [...byIcao.values()].sort((a, b) => a.icao.localeCompare(b.icao)),
    unplaced,
  };
}
