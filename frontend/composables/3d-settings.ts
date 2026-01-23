const BASE_SENSITIVITY = 0.02;

export interface SensitivitySetting {
  mouse: number; // 1–100
  movement: number; // 1–100
}

const DEFAULT_VALUE = 50;

export function useSensitivity() {
  const sensitivity = useState<SensitivitySetting>("user_sensitivity", () => ({
    mouse: DEFAULT_VALUE,
    movement: DEFAULT_VALUE,
  }));

  if (import.meta.client) {
    const mouse =
      Number(localStorage.getItem("mouse_setting")) === 0
        ? DEFAULT_VALUE
        : Number(localStorage.getItem("mouse_setting"));
    const movement =
      Number(localStorage.getItem("movement_setting")) === 0
        ? DEFAULT_VALUE
        : Number(localStorage.getItem("movement_setting"));

    if (Number.isFinite(mouse)) sensitivity.value.mouse = mouse;
    if (Number.isFinite(movement)) sensitivity.value.movement = movement;
  }

  function clamp(v: number) {
    return Math.min(100, Math.max(1, v));
  }

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
