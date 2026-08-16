export type Point = [number, number];

const RAD = Math.PI / 180;
const DEG = 180 / Math.PI;

export function greatCircle(from: Point, to: Point, steps = 64): Point[] {
  const [lat1, lon1] = [from[0] * RAD, from[1] * RAD];
  const [lat2, lon2] = [to[0] * RAD, to[1] * RAD];

  const d =
    2 *
    Math.asin(
      Math.sqrt(
        Math.sin((lat2 - lat1) / 2) ** 2 +
          Math.cos(lat1) * Math.cos(lat2) * Math.sin((lon2 - lon1) / 2) ** 2,
      ),
    );

  if (!isFinite(d) || d < 1e-9) return [from, to];

  const out: Point[] = [];
  for (let i = 0; i <= steps; i++) {
    const f = i / steps;
    const a = Math.sin((1 - f) * d) / Math.sin(d);
    const b = Math.sin(f * d) / Math.sin(d);

    const x = a * Math.cos(lat1) * Math.cos(lon1) + b * Math.cos(lat2) * Math.cos(lon2);
    const y = a * Math.cos(lat1) * Math.sin(lon1) + b * Math.cos(lat2) * Math.sin(lon2);
    const z = a * Math.sin(lat1) + b * Math.sin(lat2);

    out.push([
      Math.atan2(z, Math.sqrt(x * x + y * y)) * DEG,
      Math.atan2(y, x) * DEG,
    ]);
  }
  return unwrap(out);
}

function unwrap(points: Point[]): Point[] {
  let offset = 0;
  const out: Point[] = [points[0]];
  for (let i = 1; i < points.length; i++) {
    const prev = points[i - 1][1] + offset;
    let lon = points[i][1] + offset;
    if (lon - prev > 180) offset -= 360;
    else if (lon - prev < -180) offset += 360;
    lon = points[i][1] + offset;
    out.push([points[i][0], lon]);
  }
  return out;
}

export function distanceNm(from: Point, to: Point): number {
  const [lat1, lon1] = [from[0] * RAD, from[1] * RAD];
  const [lat2, lon2] = [to[0] * RAD, to[1] * RAD];
  const h =
    Math.sin((lat2 - lat1) / 2) ** 2 +
    Math.cos(lat1) * Math.cos(lat2) * Math.sin((lon2 - lon1) / 2) ** 2;
  return 2 * 6371 * Math.asin(Math.sqrt(h)) * 0.539957;
}
