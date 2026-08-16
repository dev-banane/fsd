import { onMounted, onUnmounted, ref } from "vue";

export function useMediaQuery(query: string) {
  const get = () =>
    typeof window !== "undefined" ? window.matchMedia(query).matches : false;

  const matches = ref(get());
  let mql: MediaQueryList | undefined;

  const update = () => {
    matches.value = mql ? mql.matches : get();
  };

  onMounted(() => {
    mql = window.matchMedia(query);
    update();
    mql.addEventListener("change", update);
  });

  onUnmounted(() => {
    mql?.removeEventListener("change", update);
  });

  return matches;
}
