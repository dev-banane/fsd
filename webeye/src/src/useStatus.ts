import { onUnmounted, ref, shallowRef } from "vue";
import { emptyStatus, fetchStatus, type Status } from "./api";

export function useStatus(intervalMs = 5000) {
  const status = shallowRef<Status>(emptyStatus());
  const connected = ref(false);
  const lastError = ref<string | null>(null);

  let timer: number | undefined;
  let controller: AbortController | undefined;

  async function tick() {
    controller?.abort();
    controller = new AbortController();
    try {
      status.value = await fetchStatus(controller.signal);
      connected.value = true;
      lastError.value = null;
    } catch (err) {
      if ((err as Error).name === "AbortError") return;
      connected.value = false;
      lastError.value = (err as Error).message;
    }
  }

  tick();
  timer = window.setInterval(tick, intervalMs);

  onUnmounted(() => {
    if (timer) window.clearInterval(timer);
    controller?.abort();
  });

  return { status, connected, lastError, refresh: tick };
}
