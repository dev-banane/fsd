import { setWorkerUrl, type StyleSpecification } from "maplibre-gl";

setWorkerUrl("/maplibre/maplibre-gl-worker.mjs");

const STYLE_URL = "https://basemaps.cartocdn.com/gl/dark-matter-gl-style/style.json";

export async function loadEnglishBasemapStyle(): Promise<StyleSpecification> {
  const res = await fetch(STYLE_URL);
  if (!res.ok) throw new Error(`basemap style fetch failed: ${res.status}`);
  const style = (await res.json()) as StyleSpecification;

  for (const layer of style.layers ?? []) {
    if (layer.type !== "symbol" || !("layout" in layer)) continue;
    const layout = layer.layout as Record<string, unknown> | undefined;
    const field = layout?.["text-field"];
    if (!field || !mentionsName(field)) continue;
    layout!["text-field"] = ["coalesce", ["get", "name_en"], ["get", "name:en"], ["get", "name"]];
  }

  return style;
}

function mentionsName(field: unknown): boolean {
  if (typeof field === "string") return field.includes("name");
  return JSON.stringify(field).includes("name");
}
