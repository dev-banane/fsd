<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { ChevronUp, Plane, RadioTower, Users } from "@lucide/vue";
import LevelControl from "./components/LevelControl.vue";
import TrafficMap from "./components/TrafficMap.vue";
import TrafficPanel from "./components/TrafficPanel.vue";
import { useMediaQuery } from "./useMediaQuery";
import { useStatus } from "./useStatus";

const { status, connected } = useStatus(5000);
const brandTitle =
  document.querySelector('meta[name="webeye-title"]')?.getAttribute("content")?.trim() ||
  "WebEye";
const selected = ref<string | null>(null);
const showPilots = ref(true);
const showSectors = ref(true);
const level = ref(100);

const isDesktop = useMediaQuery("(min-width: 768px)");
const sheetOpen = ref(false);
const sheetEl = ref<HTMLElement | null>(null);
const dragging = ref(false);
const dragY = ref(0);

const clock = computed(() => {
  const t = status.value.timestamp;
  if (!t) return "offline";
  const parts = t.trim().split(/\s+/);
  return parts.length > 1 ? parts[parts.length - 1] : t;
});

function dockOffset() {
  const el = sheetEl.value;
  if (!el) return 0;
  const chrome = el.querySelector(".we-sheet-chrome") as HTMLElement | null;
  const search = el.querySelector(".we-panel-search") as HTMLElement | null;
  const visible = (chrome?.offsetHeight ?? 0) + (search?.offsetHeight ?? 0);
  return Math.max(0, el.offsetHeight - visible);
}

function restY() {
  return sheetOpen.value ? 0 : dockOffset();
}

let startPointerY = 0;
let startY = 0;
let lastY = 0;
let lastT = 0;
let velocity = 0;
let lastToggleAt = 0;

function toggleSheet() {
  if (isDesktop.value) return;
  const now = performance.now();
  if (now - lastToggleAt < 400) return;
  lastToggleAt = now;
  sheetOpen.value = !sheetOpen.value;
}

function onSheetPointerDown(e: PointerEvent) {
  if (isDesktop.value || e.button) return;
  dragging.value = true;
  (e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
  startPointerY = e.clientY;
  startY = restY();
  dragY.value = startY;
  lastY = e.clientY;
  lastT = e.timeStamp;
  velocity = 0;
}

function onSheetPointerMove(e: PointerEvent) {
  if (!dragging.value) return;
  const max = dockOffset();
  dragY.value = Math.min(max, Math.max(0, startY + (e.clientY - startPointerY)));
  const dt = e.timeStamp - lastT;
  if (dt > 0) velocity = (e.clientY - lastY) / dt;
  lastY = e.clientY;
  lastT = e.timeStamp;
}

function onSheetPointerUp() {
  if (!dragging.value) return;
  dragging.value = false;
  const max = dockOffset();
  if (Math.abs(dragY.value - startY) < 8) {
    toggleSheet();
    return;
  }
  lastToggleAt = performance.now();
  sheetOpen.value = velocity < -0.35 || (velocity <= 0.35 && dragY.value < max * 0.5);
}

function onSelect(callsign: string | null) {
  selected.value = callsign;
  if (callsign && !isDesktop.value) sheetOpen.value = false;
}

watch(isDesktop, (desktop) => {
  dragging.value = false;
  if (desktop) sheetOpen.value = false;
});
</script>

<template>
  <div class="flex h-[100dvh] max-h-[100dvh] flex-col overflow-hidden">
    <header
      class="we-header relative z-30 flex shrink-0 items-center gap-2 border-b
             border-ink-400 bg-ink-800 md:grid md:grid-cols-[1fr_auto_1fr] md:gap-4"
    >
      <div class="flex min-w-0 items-center gap-1.5">
        <span class="truncate text-[15px] font-semibold tracking-tight">{{ brandTitle }}</span>
        <a
          href="https://github.com/dev-banane/fsd"
          target="_blank"
          rel="noopener noreferrer"
          class="flex size-8 shrink-0 items-center justify-center rounded-md text-chalk-dim
                 transition-colors duration-300 ease-[cubic-bezier(0.32,0.72,0,1)]
                 hover:text-chalk active:scale-[0.98] md:size-7"
          title="GitHub"
          aria-label="dev-banane/fsd on GitHub"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="15"
            height="15"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2.25"
            stroke-linecap="round"
            stroke-linejoin="round"
            aria-hidden="true"
          >
            <path
              d="M15 22v-4a4.8 4.8 0 0 0-1-3.5c3 0 6-2 6-5.5.08-1.25-.27-2.48-1-3.5.28-1.15.28-2.35 0-3.5 0 0-1 0-3 1.5-2.64-.5-5.36-.5-8 0C6 2 5 2 5 2c-.28 1.15-.28 2.35 0 3.5A5.403 5.403 0 0 0 4 9c0 3.5 3 5.5 6 5.5-.39.49-.68 1.05-.85 1.65-.17.6-.22 1.23-.15 1.85v4"
            />
            <path d="M9 18c-4.51 2-5-2-7-2" />
          </svg>
        </a>
      </div>

      <LevelControl
        v-model:enabled="showSectors"
        v-model:level="level"
        class="max-md:ml-auto"
      />

      <div class="flex items-center justify-end gap-3 md:gap-5">
        <div class="hidden items-center gap-4 md:flex">
          <div class="flex items-center gap-1.5" title="Controllers online">
            <RadioTower :size="13" :stroke-width="2.25" class="text-chalk-dim" />
            <span class="nums text-[13px]">{{ status.controllers.length }}</span>
          </div>
          <div class="flex items-center gap-1.5" title="Pilots online">
            <Plane :size="13" :stroke-width="2.25" class="text-chalk-dim" />
            <span class="nums text-[13px]">{{ status.pilots.length }}</span>
          </div>
          <div class="flex items-center gap-1.5" title="Connected clients">
            <Users :size="13" :stroke-width="2.25" class="text-chalk-dim" />
            <span class="nums text-[13px]">{{ status.clientCount }}</span>
          </div>
        </div>

        <span
          class="nums flex items-center gap-1.5 text-[11px]"
          :class="connected && !status.stale ? 'text-chalk-dim' : 'text-danger'"
          :title="status.timestamp ? `Data timestamp ${status.timestamp} UTC` : 'No data yet'"
        >
          <span
            class="size-1.5 rounded-full"
            :class="connected && !status.stale ? 'bg-ok' : 'bg-danger'"
          />
          <span class="md:hidden">{{ clock }}</span>
          <span class="hidden md:inline">{{ status.timestamp || "offline" }}</span>
        </span>
      </div>
    </header>

    <main class="relative flex min-h-0 flex-1">
      <div
        class="max-md:pointer-events-none max-md:absolute max-md:inset-0 max-md:z-20 md:contents"
      >
        <button
          v-show="sheetOpen && !isDesktop"
          type="button"
          class="pointer-events-auto absolute inset-0 bg-ink-900/45 md:hidden"
          aria-label="Close traffic list"
          @click="sheetOpen = false"
        />

        <aside
          ref="sheetEl"
          class="we-sheet pointer-events-auto flex flex-col border-ink-400 bg-ink-800
                 max-md:fixed max-md:inset-x-0 max-md:bottom-0
                 md:relative md:h-full md:w-[340px] md:shrink-0 md:border-r"
          :class="{ 'is-open': sheetOpen, 'is-dragging': dragging }"
          :style="dragging ? { transform: `translate3d(0, ${dragY}px, 0)` } : undefined"
        >
          <div
            class="we-sheet-chrome md:hidden"
            @pointerdown="onSheetPointerDown"
            @pointermove="onSheetPointerMove"
            @pointerup="onSheetPointerUp"
            @pointercancel="onSheetPointerUp"
          >
            <button
              type="button"
              class="flex w-full flex-col items-stretch"
              :aria-expanded="sheetOpen"
              aria-controls="we-traffic-panel"
              :aria-label="sheetOpen ? 'Collapse traffic list' : 'Expand traffic list'"
              @click="toggleSheet"
            >
              <span class="flex justify-center pt-2 pb-1">
                <span class="h-1 w-10 rounded-full bg-ink-100/80" />
              </span>
              <span class="flex min-h-11 items-center gap-3 px-4 pb-2">
                <span class="flex items-center gap-1.5 text-[12px] text-chalk-dim">
                  <RadioTower :size="13" :stroke-width="2.25" />
                  <span class="nums text-chalk">{{ status.controllers.length }}</span>
                </span>
                <span class="flex items-center gap-1.5 text-[12px] text-chalk-dim">
                  <Plane :size="13" :stroke-width="2.25" />
                  <span class="nums text-chalk">{{ status.pilots.length }}</span>
                </span>
                <span class="flex items-center gap-1.5 text-[12px] text-chalk-dim">
                  <Users :size="13" :stroke-width="2.25" />
                  <span class="nums text-chalk">{{ status.clientCount }}</span>
                </span>
                <ChevronUp
                  :size="16"
                  :stroke-width="2.25"
                  class="ml-auto text-chalk-dim transition-transform duration-300 ease-[cubic-bezier(0.32,0.72,0,1)]"
                  :class="sheetOpen ? '' : 'rotate-180'"
                />
              </span>
            </button>
          </div>

          <TrafficPanel
            id="we-traffic-panel"
            v-model:show-pilots="showPilots"
            :status="status"
            :selected="selected"
            @select="onSelect"
            @activate="sheetOpen = true"
          />
        </aside>
      </div>

      <div class="relative min-h-0 min-w-0 flex-1">
        <TrafficMap
          :status="status"
          :selected="selected"
          :show-pilots="showPilots"
          :show-sectors="showSectors"
          :level="level"
          @select="selected = $event"
        />
      </div>
    </main>
  </div>
</template>
