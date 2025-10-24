<!-- eslint-disable @typescript-eslint/no-explicit-any -->
<script lang="ts" setup>
interface ConflictItem {
  base: any;
  main: any;
  workspace: any;
}

const props = defineProps<{
  visible: boolean;
  conflicts: Record<string, Record<string, ConflictItem>>;
}>();

const emit = defineEmits<{
  (e: "close"): void;
  (e: "resolved", payload: { path: string; value: any }[]): void;
}>();

const visible = computed(() => props.visible);

// keys in deterministic order
const conflictKeys = computed(() =>
  Object.fromEntries(
    Object.keys(props.conflicts).map((camId) => {
      return [camId, Object.keys(props.conflicts[camId] || {}).sort()];
    }),
  ),
);

// local selection state: 'main' | 'workspace' | 'manual'
const selected = reactive<Record<string, "main" | "workspace" | "manual">>({});
const manualEdits = reactive<Record<string, string>>({});
const manualErrors = reactive<Record<string, string | null>>({});

// initialize defaults whenever conflicts change
watch(
  () => props.conflicts,
  (newVal) => {
    for (const key of Object.keys(newVal || {})) {
      // default pick: if Main equals Workspace -> pick it; else prefer Workspace
      const item = newVal[key];
      if (deepEqual(item?.Main, item?.Workspace)) {
        selected[key] = "main";
      } else {
        selected[key] = "workspace";
      }
      manualEdits[key] = "";
      manualErrors[key] = null;
    }
  },
  { immediate: true },
);

function close() {
  emit("close");
}

function select(key: string, which: "main" | "workspace" | "manual") {
  selected[key] = which;
  manualErrors[key] = null;
  if (which !== "manual") manualEdits[key] = "";
}

function pretty(v: any) {
  try {
    return JSON.stringify(v, null, 2);
  } catch {
    return String(v);
  }
}

function buttonClass(active: boolean) {
  return [
    "px-3 py-1 rounded text-sm border",
    active ? "bg-gray-100 border-gray-300" : "bg-white border-gray-200",
  ].join(" ");
}

function deepEqual(a: any, b: any) {
  return JSON.stringify(a) === JSON.stringify(b);
}

function parseManual(key: string): { ok: boolean; value?: any; err?: string } {
  const raw = manualEdits[key];
  if (!raw || raw.trim() === "")
    return { ok: false, err: "Empty manual value" };
  try {
    const parsed = JSON.parse(raw);
    return { ok: true, value: parsed };
  } catch (e: any) {
    return { ok: false, err: e.message || "Invalid JSON" };
  }
}

function applyAll() {
  const results: { path: string; value: any }[] = [];
  let hasError = false;

  for (const camId in conflictKeys.value) {
    for (const key of conflictKeys.value[camId]!) {
      const choice = selected[key];
      const item = props.conflicts[key];

      if (choice === "main") {
        results.push({ path: key, value: item?.Main });
      } else if (choice === "workspace") {
        results.push({ path: key, value: item?.Workspace });
      } else if (choice === "manual") {
        const parsed = parseManual(key);
        if (!parsed.ok) {
          manualErrors[key] = parsed.err || "Invalid JSON";
          hasError = true;
        } else {
          manualErrors[key] = null;
          results.push({ path: key, value: parsed.value });
        }
      }
    }
  }

  if (hasError) {
    // keep dialog open and show errors
    return;
  }

  emit("resolved", results);
  emit("close");
}
</script>

<template>
  <transition name="fade">
    <div
      v-if="visible"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
    >
      <div
        class="bg-white rounded-2xl shadow-xl w-full max-w-4xl mx-4 overflow-hidden"
      >
        <header class="flex items-center justify-between px-6 py-4 border-b">
          <h3 class="text-lg font-semibold">Resolve Merge Conflicts</h3>
          <button class="text-gray-500 hover:text-gray-700" @click="close">
            ✕
          </button>
        </header>

        <main class="p-6 space-y-4 max-h-[70vh] overflow-auto">
          <p class="text-sm text-gray-600">
            Choose how to resolve each conflicting field. You can pick
            <strong>Main</strong>, <strong>Workspace</strong>, or edit manually.
          </p>

          <div
            v-if="Object.keys(conflictKeys).length === 0"
            class="text-center py-12 text-gray-500"
          >
            No conflicts to resolve.
          </div>

          <div
            v-for="(conflictsOfCam, camId) in conflictKeys"
            :key="camId"
            class="border rounded-lg p-4"
          >
            <div
              v-for="key in conflictsOfCam"
              :key="key"
              class="border rounded-lg p-4"
            >
              <div class="flex items-start justify-between">
                <div>
                  <div class="text-sm text-gray-700 font-medium">{{ key }}</div>
                  <div class="text-xs text-gray-500">Field path</div>
                </div>
                <div class="flex items-center gap-2">
                  <button
                    :class="buttonClass(selected[key] === 'main')"
                    title="Use Main's version"
                    @click="select(key, 'main')"
                  >
                    Main
                  </button>
                  <button
                    :class="buttonClass(selected[key] === 'workspace')"
                    title="Use Workspace's version"
                    @click="select(key, 'workspace')"
                  >
                    Workspace
                  </button>
                  <button
                    :class="buttonClass(selected[key] === 'manual')"
                    title="Edit manually"
                    @click="select(key, 'manual')"
                  >
                    Manual
                  </button>
                </div>
              </div>

              <div class="mt-3 grid grid-cols-3 gap-3 text-xs">
                <div class="p-2 border rounded">
                  <div class="font-semibold text-emerald-600">Base</div>
                  <pre class="whitespace-pre-wrap text-[12px] mt-2">{{
                    pretty(props.conflicts[camId]?.[key]?.base)
                  }}</pre>
                </div>

                <div class="p-2 border rounded">
                  <div class="font-semibold text-blue-600">Main</div>
                  <pre class="whitespace-pre-wrap text-[12px] mt-2">{{
                    pretty(props.conflicts[camId]?.[key]?.main)
                  }}</pre>
                </div>

                <div class="p-2 border rounded">
                  <div class="font-semibold text-purple-600">Workspace</div>
                  <pre class="whitespace-pre-wrap text-[12px] mt-2">{{
                    pretty(props.conflicts[camId]?.[key]?.workspace)
                  }}</pre>
                </div>
              </div>

              <div v-if="selected[key] === 'manual'" class="mt-3">
                <label class="block text-xs font-medium text-gray-600"
                  >Manual value (JSON)</label
                >
                <textarea
                  v-model="manualEdits[key]"
                  rows="4"
                  class="mt-1 block w-full border rounded p-2 text-sm font-mono"
                ></textarea>
                <div class="mt-2 text-xs text-gray-500">
                  Enter a JSON value. Example: <code>true</code>,
                  <code>"string"</code>, <code>{"x":1}</code>.
                </div>
                <div v-if="manualErrors[key]" class="text-xs text-red-600 mt-1">
                  {{ manualErrors[key] }}
                </div>
              </div>

              <div v-else class="mt-3 text-xs text-gray-600">
                Selected: <strong>{{ selected[key] }}</strong>
              </div>
            </div>
          </div>
        </main>

        <footer class="flex items-center justify-end gap-3 px-6 py-4 border-t">
          <button class="px-4 py-2 rounded bg-white border" @click="close">
            Cancel
          </button>
          <button
            class="px-4 py-2 rounded bg-blue-600 text-white"
            @click="applyAll"
          >
            Apply Resolutions
          </button>
        </footer>
      </div>
    </div>
  </transition>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
