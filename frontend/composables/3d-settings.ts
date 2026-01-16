const BASE_SENSITIVITY = 0.02;

export interface SensitivitySetting {
  mouse: number; // 1–100
  movement: number; // 1–100
}

const DEFAULT_VALUE = 50;

function getStoredNumber(key: string, fallback = DEFAULT_VALUE): number {
  if (import.meta.client) {
    const value = localStorage.getItem(key);
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : fallback;
  }
  return fallback;
}

export function useSensitivity() {
  const sensitivity = useState<SensitivitySetting>("user_sensitivity", () => ({
    mouse: getStoredNumber("mouse_setting"),
    movement: getStoredNumber("movement_setting"),
  }));

  onMounted(() => {
    const mouse = Number(localStorage.getItem("mouse_setting"));
    const movement = Number(localStorage.getItem("movement_setting"));

    if (Number.isFinite(mouse)) sensitivity.value.mouse = mouse;
    if (Number.isFinite(movement)) sensitivity.value.movement = movement;
  });

  function setMouse(value: number) {
    const v = clamp(value);
    sensitivity.value.mouse = v;
    localStorage.setItem("mouse_setting", String(v));
  }

  function setMovement(value: number) {
    const v = clamp(value);
    sensitivity.value.movement = v;
    localStorage.setItem("movement_setting", String(v));
  }

  function clamp(v: number) {
    return Math.min(100, Math.max(1, v));
  }

  const normalizedSensitivity = computed(() => ({
    mouse: sensitivity.value.mouse * BASE_SENSITIVITY,
    movement: sensitivity.value.movement * BASE_SENSITIVITY,
  }));

  return {
    sensitivity,
    normalizedSensitivity,
    setMouse,
    setMovement,
  };
}
