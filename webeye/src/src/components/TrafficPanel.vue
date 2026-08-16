<script setup lang="ts">
import { computed, ref } from "vue";
import { ChevronRight } from "@lucide/vue";
import { flightLevel, positionColor, positionRank, type Controller, type Status } from "../api";
import { buildStations, isSectorPosition, AIRPORT_POSITIONS, PILL_ORDER, positionLetter } from "../stations";

const props = defineProps<{ status: Status; selected: string | null }>();
const emit = defineEmits<{
  select: [callsign: string | null];
  "update:showPilots": [value: boolean];
}>();

const showPilots = defineModel<boolean>("showPilots", { default: true });
const query = ref("");
const pilotsOpen = ref(false);

const matches = (fields: string[]) =>
  !query.value ||
  fields.some((f) => f?.toLowerCase().includes(query.value.toLowerCase()));

const filteredControllers = computed(() =>
  props.status.controllers.filter((c) => matches([c.callsign, c.name, c.frequency])),
);

const sectors = computed(() =>
  filteredControllers.value
    .filter(isSectorPosition)
    .sort(
      (a, b) =>
        positionRank[a.position] - positionRank[b.position] ||
        a.callsign.localeCompare(b.callsign),
    ),
);

const airports = computed(() => {
  const { stations, unplaced } = buildStations(filteredControllers.value);
  const sortAtAirport = (a: Controller, b: Controller) => {
    const ia = PILL_ORDER.indexOf(a.position);
    const ib = PILL_ORDER.indexOf(b.position);
    return (ia === -1 ? 99 : ia) - (ib === -1 ? 99 : ib) || a.callsign.localeCompare(b.callsign);
  };
  return {
    stations: stations.map((s) => ({
      ...s,
      controllers: [...s.controllers].sort(sortAtAirport),
    })),
    unplaced: [
      ...unplaced.filter((c) => matches([c.callsign, c.name, c.frequency])),
      ...filteredControllers.value.filter(
        (c) => !isSectorPosition(c) && !AIRPORT_POSITIONS.includes(c.position),
      ),
    ].sort((a, b) => a.callsign.localeCompare(b.callsign)),
  };
});

const pilots = computed(() =>
  props.status.pilots
    .filter((p) =>
      matches([p.callsign, p.name, p.type, p.departure_ap, p.arrival_ap]),
    )
    .sort((a, b) => a.callsign.localeCompare(b.callsign)),
);

const pick = (cs: string) => emit("select", props.selected === cs ? null : cs);
</script>

<template>
  <aside class="flex w-[340px] shrink-0 flex-col border-r border-ink-400 bg-ink-800">
    <div class="p-3 pb-2">
      <input
        v-model="query"
        type="search"
        placeholder="Filter callsign, name, frequency…"
        class="w-full rounded-lg border border-ink-400 bg-ink-900 px-3 py-1.5
               text-[12px] text-chalk placeholder:text-chalk-dim/50
               focus:border-brand-stroke focus:outline-none"
      />
    </div>

    <div class="min-h-0 flex-1 overflow-y-auto pb-3">
      <h2
        class="px-3 pt-1 pb-1.5 text-[10px] font-semibold uppercase tracking-wider text-chalk-dim/70"
      >
        Controllers
        <span class="nums ml-1">{{ status.controllers.length }}</span>
      </h2>

      <p
        v-if="!filteredControllers.length"
        class="px-3 py-5 text-center text-[12px] text-chalk-dim/60"
      >
        No controllers online.
      </p>

      <template v-else>
        <section v-if="sectors.length">
          <h3
            class="px-3 pb-1 text-[10px] font-semibold uppercase tracking-wider text-chalk-dim/50"
          >
            Sectors
          </h3>
          <button
            v-for="c in sectors"
            :key="c.callsign"
            class="we-atc-row w-full text-left"
            :class="selected === c.callsign ? 'bg-ink-500' : ''"
            @click="pick(c.callsign)"
          >
            <div class="we-atc-main">
              <span class="we-facility" :style="{ background: positionColor[c.position] }">
                {{ c.position }}
              </span>
              <span class="we-atc-callsign">{{ c.callsign }}</span>
              <span class="we-atc-freq nums">{{ c.frequency }}</span>
            </div>
            <div class="we-atc-name">{{ c.name }}</div>
          </button>
        </section>

        <section v-for="station in airports.stations" :key="station.icao">
          <h3
            class="px-3 pt-2 pb-1 text-[10px] font-semibold uppercase tracking-wider text-chalk-dim/50"
          >
            {{ station.icao }}
            <span class="ml-1 inline-flex align-middle">
              <span
                v-for="p in station.positions"
                :key="p"
                class="we-station-pill"
                :style="{ background: positionColor[p] }"
              >{{ positionLetter[p] }}</span>
            </span>
          </h3>
          <button
            v-for="c in station.controllers"
            :key="c.callsign"
            class="we-atc-row w-full text-left"
            :class="selected === c.callsign ? 'bg-ink-500' : ''"
            @click="pick(c.callsign)"
          >
            <div class="we-atc-main">
              <span class="we-facility" :style="{ background: positionColor[c.position] }">
                {{ c.position }}
              </span>
              <span class="we-atc-callsign">{{ c.callsign }}</span>
              <span class="we-atc-freq nums">{{ c.frequency }}</span>
            </div>
            <div class="we-atc-name">{{ c.name }}</div>
          </button>
        </section>

        <section v-if="airports.unplaced.length">
          <h3
            class="px-3 pt-2 pb-1 text-[10px] font-semibold uppercase tracking-wider text-chalk-dim/50"
          >
            Other
          </h3>
          <button
            v-for="c in airports.unplaced"
            :key="c.callsign"
            class="we-atc-row w-full text-left"
            :class="selected === c.callsign ? 'bg-ink-500' : ''"
            @click="pick(c.callsign)"
          >
            <div class="we-atc-main">
              <span class="we-facility" :style="{ background: positionColor[c.position] }">
                {{ c.position }}
              </span>
              <span class="we-atc-callsign">{{ c.callsign }}</span>
              <span class="we-atc-freq nums">{{ c.frequency }}</span>
            </div>
            <div class="we-atc-name">{{ c.name }}</div>
          </button>
        </section>
      </template>

      <div class="mt-4 flex items-center gap-2 px-3 pb-1.5">
        <button
          class="flex items-center gap-1 text-[10px] font-semibold uppercase tracking-wider text-chalk-dim/70 hover:text-chalk"
          @click="pilotsOpen = !pilotsOpen"
        >
          <ChevronRight
            :size="13"
            :stroke-width="2.5"
            class="transition-transform"
            :class="pilotsOpen ? 'rotate-90' : ''"
          />
          Pilots
          <span class="nums">{{ status.pilots.length }}</span>
        </button>

        <label class="ml-auto flex cursor-pointer items-center gap-1.5 text-[10px] text-chalk-dim">
          <input v-model="showPilots" type="checkbox" class="accent-brand" />
          on map
        </label>
      </div>

      <template v-if="pilotsOpen">
        <p
          v-if="!pilots.length"
          class="px-3 py-5 text-center text-[12px] text-chalk-dim/60"
        >
          No pilots connected.
        </p>
        <button
          v-for="p in pilots"
          :key="p.callsign"
          class="we-atc-row w-full text-left"
          :class="selected === p.callsign ? 'bg-ink-500' : ''"
          @click="pick(p.callsign)"
        >
          <div class="we-atc-main">
            <span class="we-atc-callsign">{{ p.callsign }}</span>
            <span class="we-atc-freq nums">{{ flightLevel(p.alt) }}</span>
          </div>
          <div class="we-atc-name">
            {{ p.departure_ap || "????" }}
            →
            {{ p.arrival_ap || "????" }}
            <span class="nums ml-2 opacity-70">{{ p.type || "—" }}</span>
          </div>
        </button>
      </template>
    </div>
  </aside>
</template>
