<script setup lang="ts">
import { ref } from "vue";
import { RadioTower, Plane, Users } from "@lucide/vue";
import LevelControl from "./components/LevelControl.vue";
import TrafficMap from "./components/TrafficMap.vue";
import TrafficPanel from "./components/TrafficPanel.vue";
import { useStatus } from "./useStatus";

const { status, connected } = useStatus(5000);
const selected = ref<string | null>(null);
const showPilots = ref(true);
const showSectors = ref(true);
const level = ref(100);
</script>

<template>
  <div class="flex h-full flex-col">
    <header
      class="grid shrink-0 grid-cols-[1fr_auto_1fr] items-center gap-4 border-b
             border-ink-400 bg-ink-800 px-4 py-2.5"
    >
      <div class="flex items-baseline gap-2">
        <span class="text-[15px] font-semibold tracking-tight">WebEye</span>
        <span class="text-[11px] text-chalk-dim/60">FSD live traffic</span>
      </div>

      <LevelControl v-model:enabled="showSectors" v-model:level="level" />

      <div class="flex items-center justify-end gap-5">
        <div class="hidden items-center gap-4 sm:flex">
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
          class="nums text-[11px]"
          :class="connected && !status.stale ? 'text-chalk-dim' : 'text-danger'"
          :title="status.timestamp ? `Data timestamp ${status.timestamp} UTC` : 'No data yet'"
        >
          {{ status.timestamp || "offline" }}
        </span>
      </div>
    </header>

    <main class="flex min-h-0 flex-1">
      <TrafficPanel
        v-model:show-pilots="showPilots"
        :status="status"
        :selected="selected"
        @select="selected = $event"
      />
      <div class="relative min-w-0 flex-1">
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
