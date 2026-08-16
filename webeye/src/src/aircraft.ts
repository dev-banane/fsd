import iconMap from "./assets/aircraft-icons.json";

const MAP = iconMap as Record<string, string>;
const FALLBACK = "a320";

export function iconFor(type: string): string {
  if (!type) return FALLBACK;
  const parts = type.toUpperCase().split("/").filter(Boolean);
  for (const part of parts.length > 1 ? parts.slice(1).concat(parts[0]) : parts) {
    const code = part.replace(/[^A-Z0-9]/g, "");
    if (MAP[code]) return MAP[code];
    if (code.startsWith("P28") && MAP["P28A"]) return MAP["P28A"];
  }
  return FALLBACK;
}

const cache = new Map<string, Promise<string>>();

function loadSvg(icon: string): Promise<string> {
  let pending = cache.get(icon);
  if (!pending) {
    pending = fetch(`/aircraft/${icon}.svg`)
      .then((r) => (r.ok ? r.text() : Promise.reject(new Error(String(r.status)))))
      .catch(() => "");
    cache.set(icon, pending);
  }
  return pending;
}

export async function aircraftMarkup(
  type: string,
  heading: number,
  color: string,
  size: number,
): Promise<string> {
  const svg = await loadSvg(iconFor(type));
  if (!svg) {
    return `<div style="width:${size}px;height:${size}px"></div>`;
  }

  const tinted = svg
    .replace(/fill="white"/gi, `fill="${color}"`)
    .replace(/<svg([^>]*)>/i, (_m, attrs: string) => {
      const cleaned = attrs
        .replace(/\swidth="[^"]*"/i, "")
        .replace(/\sheight="[^"]*"/i, "");
      return `<svg${cleaned} width="${size}" height="${size}" preserveAspectRatio="xMidYMid meet">`;
    });

  return `<div style="width:${size}px;height:${size}px;transform:rotate(${heading}deg)">${tinted}</div>`;
}

export const aircraftSize = (zoom: number): number => {
  const coarse =
    typeof window !== "undefined" && window.matchMedia("(pointer: coarse)").matches;
  const min = coarse ? 22 : 18;
  const max = coarse ? 44 : 40;
  return Math.round(Math.max(min, Math.min(max, 12 + zoom * 2.2)));
};
