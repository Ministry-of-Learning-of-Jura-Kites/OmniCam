export interface SensitivitySetting {
  mouse: number; // Mouse sensitivity 1–100
  movement: number; // Movement sensitivity 1–100
}

export function useSensitivity() {
  const sensitivity = useState<SensitivitySetting>("user_sensitivity", () => ({
    mouse: 50,
    movement: 50,
  }));

  function setMouse(value: number) {
    sensitivity.value.mouse = clamp(value);
  }

  function setMovement(value: number) {
    sensitivity.value.movement = clamp(value);
  }

  function clamp(v: number) {
    return Math.min(100, Math.max(1, v));
  }

  return {
    sensitivity,
    setMouse,
    setMovement,
  };
}
