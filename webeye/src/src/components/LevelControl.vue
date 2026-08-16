<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from "vue";
import { Layers } from "@lucide/vue";
import { useMediaQuery } from "../useMediaQuery";

const enabled = defineModel<boolean>("enabled", { default: true });
const level = defineModel<number>("level", { default: 100 });

const isDesktop = useMediaQuery("(min-width: 768px)");
const menuOpen = ref(false);
const btn = ref<HTMLElement | null>(null);
const menuTop = ref(0);

const PRESETS = [
  { label: "GND", fl: 0 },
  { label: "FL100", fl: 100 },
  { label: "FL245", fl: 245 },
  { label: "FL350", fl: 350 },
];

function clamp(value: number) {
  return Math.max(0, Math.min(660, Math.round(value)));
}

const levelLabel = computed(() =>
  level.value === 0 ? "GND" : `FL${String(level.value).padStart(3, "0")}`,
);

function placeMenu() {
  const r = btn.value?.getBoundingClientRect();
  menuTop.value = (r?.bottom ?? 48) + 8;
}

function onButtonClick() {
  if (isDesktop.value) {
    enabled.value = !enabled.value;
    return;
  }
  menuOpen.value = !menuOpen.value;
  if (menuOpen.value) placeMenu();
}

function onDocPointer(e: PointerEvent) {
  if (!menuOpen.value) return;
  const t = e.target as Node;
  if (btn.value?.contains(t)) return;
  const menu = document.getElementById("we-level-menu");
  if (menu?.contains(t)) return;
  menuOpen.value = false;
}

function onKey(e: KeyboardEvent) {
  if (e.key === "Escape") menuOpen.value = false;
}

onMounted(() => {
  document.addEventListener("pointerdown", onDocPointer);
  window.addEventListener("keydown", onKey);
  window.addEventListener("resize", placeMenu);
});

onUnmounted(() => {
  document.removeEventListener("pointerdown", onDocPointer);
  window.removeEventListener("keydown", onKey);
  window.removeEventListener("resize", placeMenu);
});

watch(isDesktop, (desktop) => {
  if (desktop) menuOpen.value = false;
});
</script>

<template>
  <div class="relative flex items-center gap-3">
    <button
      ref="btn"
      type="button"
      class="flex min-h-11 min-w-11 items-center gap-1.5 rounded-md px-1.5 py-1
             text-[11px] font-medium tracking-tight transition-colors
             duration-300 ease-[cubic-bezier(0.32,0.72,0,1)] active:scale-[0.98]
             md:min-h-0 md:min-w-0"
      :class="enabled ? 'text-chalk' : 'text-chalk-dim hover:text-chalk'"
      :title="isDesktop ? (enabled ? 'Hide sectors' : 'Show sectors') : 'Sector altitude'"
      :aria-expanded="!isDesktop && menuOpen"
      aria-controls="we-level-menu"
      @click="onButtonClick"
    >
      <Layers :size="14" :stroke-width="2.25" />
      <span class="nums">{{ levelLabel }}</span>
    </button>

    <input
      :value="level"
      type="range"
      min="0"
      max="660"
      step="5"
      class="hidden h-1 w-32 accent-brand md:block"
      :disabled="!enabled"
      :class="enabled ? '' : 'opacity-40'"
      @input="level = clamp(+($event.target as HTMLInputElement).value)"
    />

    <div class="hidden items-center gap-1 md:flex">
      <button
        v-for="p in PRESETS"
        :key="p.label"
        type="button"
        class="nums rounded-md px-1.5 py-0.5 text-[10px] transition-colors
               duration-300 ease-[cubic-bezier(0.32,0.72,0,1)]"
        :class="[
          !enabled && 'opacity-40',
          level === p.fl
            ? 'bg-ink-500 text-chalk'
            : 'text-chalk-dim hover:bg-ink-600 hover:text-chalk',
        ]"
        :disabled="!enabled"
        @click="level = p.fl"
      >
        {{ p.label }}
      </button>
    </div>

    <Teleport to="body">
      <div
        v-if="menuOpen && !isDesktop"
        id="we-level-menu"
        class="fixed inset-x-3 z-40 rounded-xl border border-ink-400 bg-ink-800 p-3
               shadow-[0_16px_40px_-16px_rgba(0,0,0,0.55)]"
        :style="{ top: `${menuTop}px` }"
        role="dialog"
        aria-label="Sector altitude"
      >
        <div class="mb-3 flex items-center justify-between gap-3">
          <span class="text-[12px] text-chalk-dim">Sectors on map</span>
          <button
            type="button"
            class="rounded-full px-3 py-1 text-[11px] font-medium transition-colors
                   duration-300 ease-[cubic-bezier(0.32,0.72,0,1)] active:scale-[0.98]"
            :class="enabled ? 'bg-ink-500 text-chalk' : 'bg-ink-900 text-chalk-dim'"
            @click="enabled = !enabled"
          >
            {{ enabled ? "On" : "Off" }}
          </button>
        </div>

        <input
          :value="level"
          type="range"
          min="0"
          max="660"
          step="5"
          class="h-8 w-full accent-brand"
          :disabled="!enabled"
          :class="enabled ? '' : 'opacity-40'"
          @input="level = clamp(+($event.target as HTMLInputElement).value)"
        />

        <div class="mt-2 flex items-center justify-between gap-1">
          <button
            v-for="p in PRESETS"
            :key="p.label"
            type="button"
            class="nums min-h-11 flex-1 rounded-md px-1.5 text-[11px] transition-colors
                   duration-300 ease-[cubic-bezier(0.32,0.72,0,1)] active:scale-[0.98]"
            :class="[
              !enabled && 'opacity-40',
              level === p.fl
                ? 'bg-ink-500 text-chalk'
                : 'text-chalk-dim hover:bg-ink-600 hover:text-chalk',
            ]"
            :disabled="!enabled"
            @click="level = p.fl"
          >
            {{ p.label }}
          </button>
        </div>
      </div>
    </Teleport>
  </div>
</template>
