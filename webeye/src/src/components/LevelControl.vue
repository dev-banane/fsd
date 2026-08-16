<script setup lang="ts">
import { Layers } from "@lucide/vue";

const enabled = defineModel<boolean>("enabled", { default: true });
const level = defineModel<number>("level", { default: 100 });

const PRESETS = [
  { label: "GND", fl: 0 },
  { label: "FL100", fl: 100 },
  { label: "FL245", fl: 245 },
  { label: "FL350", fl: 350 },
];

function clamp(value: number) {
  return Math.max(0, Math.min(660, Math.round(value)));
}
</script>

<template>
  <div class="flex items-center gap-3">
    <button
      class="flex items-center gap-1.5 rounded-md px-1.5 py-1 text-[11px] font-medium
             tracking-tight transition-colors"
      :class="enabled ? 'text-chalk' : 'text-chalk-dim hover:text-chalk'"
      :title="enabled ? 'Hide sectors' : 'Show sectors'"
      @click="enabled = !enabled"
    >
      <Layers :size="14" :stroke-width="2.25" />
      <span class="nums">
        {{ level === 0 ? "GND" : `FL${String(level).padStart(3, "0")}` }}
      </span>
    </button>

    <input
      :value="level"
      type="range"
      min="0"
      max="660"
      step="5"
      class="h-1 w-32 accent-brand"
      :disabled="!enabled"
      :class="enabled ? '' : 'opacity-40'"
      @input="level = clamp(+($event.target as HTMLInputElement).value)"
    />

    <div class="hidden items-center gap-1 md:flex">
      <button
        v-for="p in PRESETS"
        :key="p.label"
        class="nums rounded-md px-1.5 py-0.5 text-[10px] transition-colors"
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
</template>
