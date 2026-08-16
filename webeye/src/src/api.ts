export type Position = "DEL" | "GND" | "TWR" | "APP" | "CTR" | "ATIS" | "OBS";

export interface Controller {
  callsign: string;
  cid: string;
  name: string;
  lat: number | null;
  lon: number | null;
  server: string;
  rating: number;
  logon: string;
  frequency: string;
  range: number;
  facility: number;
  position: Position;
  airport: string;
  airport_pos?: LatLon;
}

export interface LatLon {
  lat: number;
  lon: number;
}

export interface Pilot {
  callsign: string;
  cid: string;
  name: string;
  lat: number | null;
  lon: number | null;
  server: string;
  rating: number;
  logon: string;
  alt: number;
  groundspeed: number;
  type: string;
  departure_ap: string;
  arrival_ap: string;
  sqwk: string;
  crzlvl: number;
  tas: number;
  flight_rules: string;
  heading: number;
  pitch: number;
  bank: number;
  onground: boolean;
  departure_pos?: LatLon;
  arrival_pos?: LatLon;
}

export interface ServerInfo {
  ident: string;
  hostname: string;
  location: string;
  name: string;
}

export interface Sector {
  callsign: string;
  position: Position;
  name: string;
  color: string;
  min: number;
  max: number;
  polygon: [number, number][];
}

export interface Status {
  timestamp: string;
  clientCount: number;
  controllers: Controller[];
  pilots: Pilot[];
  servers: ServerInfo[];
  sectors: Sector[];
  history: Record<string, [number, number][]>;
  stale: boolean;
}

export const emptyStatus = (): Status => ({
  timestamp: "",
  clientCount: 0,
  controllers: [],
  pilots: [],
  servers: [],
  sectors: [],
  history: {},
  stale: true,
});

export async function fetchStatus(signal?: AbortSignal): Promise<Status> {
  const res = await fetch("/fsd-status", { cache: "no-store", signal });
  if (!res.ok) throw new Error(`fsd-status responded ${res.status}`);
  return (await res.json()) as Status;
}

export const positionColor: Record<Position, string> = {
  DEL: "var(--color-pos-del)",
  GND: "var(--color-pos-gnd)",
  TWR: "var(--color-pos-twr)",
  APP: "var(--color-pos-app)",
  CTR: "var(--color-pos-ctr)",
  ATIS: "var(--color-pos-atis)",
  OBS: "var(--color-chalk-dim)",
};

export const positionRank: Record<Position, number> = {
  CTR: 0,
  APP: 1,
  ATIS: 2,
  TWR: 3,
  GND: 4,
  DEL: 5,
  OBS: 6,
};

export const flightLevel = (feet: number): string =>
  feet > 0 ? `FL${String(Math.round(feet / 100)).padStart(3, "0")}` : "—";

export const formatAlt = (feet: number): string =>
  feet ? `${feet.toLocaleString("en-US")} ft` : "—";
