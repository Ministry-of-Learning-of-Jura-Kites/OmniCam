<script setup lang="ts">
import { ref, reactive, watch, computed, onMounted } from "vue";
import SuccessDialog from "~/components/dialog/SuccessDialog.vue";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
} from "@/components/ui/select";

const config = useRuntimeConfig();

const props = defineProps<{
  open: boolean;
  projectId: string;
  initialSelected?: { userId: string; role: string }[];
}>();

const emit = defineEmits<{
  (e: "update:open", value: boolean): void;
  (
    e: "submit",
    data: { projectId: string; userId: string; role: string }[],
  ): void;
  (e: "members-added"): void;
}>();

type UserItem = {
  id: string;
  username: string;
  email: string | null;
  first_name: string | null;
  last_name: string | null;
  role?: string | null;
};

type SelectedEntry = {
  user: UserItem;
  role: "owner" | "project_manager" | "collaborator";
};
const isSuccessDialogOpen = ref(false);
const successMessage = ref("");
const page = ref(1);
const pageSize = ref(5);
const total = ref(0);
const users = ref<UserItem[]>([]);
const searchText = ref("");
const loading = ref(false);
const debounceTimer = ref<number | null>(null);
const selected = reactive<Record<string, SelectedEntry>>({});
const globalRole = ref<"owner" | "project_manager" | "collaborator">(
  "collaborator",
);

async function fetchUsers() {
  loading.value = true;
  try {
    const res = await $fetch<{ data: UserItem[]; count: number }>(
      `http://${config.public.NUXT_PUBLIC_BACKEND_HOST}/api/v1/projects/${props.projectId}/userForAddMembers`,
      {
        method: "GET",
        query: {
          pageSize: pageSize.value,
          page: page.value,
          search: searchText.value,
        },
        credentials: "include",
      },
    );

    const fetched = res.data.map((d) => ({
      id: d.id,
      username: d.username,
      email: d.email ?? null,
      first_name: d.first_name ?? null,
      last_name: d.last_name ?? null,
      role: d.role ?? null,
    }));

    total.value = res.count ?? 0;

    const selectedIds = Object.keys(selected);
    const merged = fetched.map((u) => {
      if (selectedIds.includes(u.id)) {
        return selected[u.id]!.user;
      }
      return u;
    });

    const missingSelected = selectedIds
      .filter((id) => !merged.some((u) => u?.id === id))
      .map((id) => selected[id]?.user);

    users.value = [
      ...missingSelected.filter((u): u is UserItem => u !== undefined),
      ...merged,
    ];
  } catch (err) {
    console.error("fetchUsers error", err);
  } finally {
    loading.value = false;
  }
}
async function addMembers(out: { userId: string; role: string }[]) {
  console.log("Submitting selected users:", out);
  try {
    const res = await $fetch(
      `http://${config.public.NUXT_PUBLIC_BACKEND_HOST}/api/v1/projects/${props.projectId}/members`,
      {
        method: "POST",
        body: out,
        credentials: "include",
      },
    );

    console.log("Add members response:", res);

    // success feedback
    successMessage.value = "Add Member Success.";
    isSuccessDialogOpen.value = true;
    emit("members-added");
    emit("update:open", false);
  } catch (err: unknown) {
    console.error("Error adding members:", err);

    const errorMessage =
      (err as { data?: { error?: string } })?.data?.error ||
      "Add members failed";
    throw new Error(errorMessage);
  }
}

function triggerFetchDebounced() {
  if (debounceTimer.value) clearTimeout(debounceTimer.value);
  debounceTimer.value = window.setTimeout(async () => {
    page.value = 1;
    await fetchUsers();
    debounceTimer.value = null;
  }, 350);
}

watch(searchText, triggerFetchDebounced);
watch(page, fetchUsers);

watch(
  () => props.open,
  async (val) => {
    if (val) {
      page.value = 1;
      searchText.value = "";
      await fetchUsers();
    } else {
      for (const key in selected) {
        delete selected[key];
      }
    }
  },
);

function toggleUserSelection(u: UserItem, checked: boolean) {
  if (checked) {
    selected[u.id] = {
      user: u,
      role: selected[u.id]?.role ?? globalRole.value,
    };
  } else {
    delete selected[u.id];
  }
}

function isChecked(u: UserItem) {
  return selected[u.id] != undefined;
}

function setUserRole(userId: string, role: SelectedEntry["role"]) {
  if (selected[userId]) selected[userId].role = role;
}

function applyGlobalRole() {
  for (const id in selected) {
    if (selected[id]) {
      selected[id].role = globalRole.value;
    }
  }
}

const selectedList = computed(() =>
  Object.values(selected).map((s) => ({
    userId: s.user.id,
    role: s.role,
    displayName: s.user.username,
    email: s.user.email,
  })),
);

const totalPages = computed(() =>
  Math.max(1, Math.ceil(total.value / pageSize.value)),
);

function goPage(p: number) {
  if (p < 1) p = 1;
  if (p > totalPages.value) p = totalPages.value;
  page.value = p;
}

function handleConfirm() {
  const out = Object.values(selected).map((s) => ({
    userId: s.user.id,
    role: s.role,
  }));

  addMembers(out);
}

function handleCancel() {
  for (const key in selected) delete selected[key];
  emit("update:open", false);
}

onMounted(fetchUsers);
</script>

<template>
  <Dialog
    :open="props.open"
    @update:open="(v: boolean) => emit('update:open', v)"
  >
    <DialogContent class="max-w-3xl w-full">
      <DialogHeader>
        <DialogTitle>Add Users to Project</DialogTitle>
      </DialogHeader>

      <div class="mt-2">
        <Input v-model="searchText" placeholder="Search users..." />
      </div>

      <div class="mt-3">
        <p class="text-sm font-medium mb-2">Selected Users</p>
        <div class="flex flex-col gap-2">
          <template v-if="selectedList.length === 0">
            <p class="text-sm text-muted-foreground">No users selected</p>
          </template>
          <template v-else>
            <div
              v-for="s in selectedList"
              :key="s.userId"
              class="flex items-center justify-between bg-slate-50 border rounded px-3 py-1"
            >
              <div class="text-sm">
                <span class="font-medium text-gray-500 ml-2">{{
                  s.displayName
                }}</span>
                <span class="ml-3 text-xs text-gray-600">({{ s.role }})</span>
              </div>
            </div>
          </template>
        </div>
      </div>

      <div class="mt-3 flex items-center gap-3">
        <p class="text-sm font-medium">Apply role to selected:</p>
        <Select v-model="globalRole" class="w-56">
          <SelectTrigger>
            <SelectValue placeholder="Select role" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="project_manager">Project Manager</SelectItem>
            <SelectItem value="collaborator">Collaborator</SelectItem>
          </SelectContent>
        </Select>
        <Button
          variant="outline"
          size="sm"
          class="ml-2"
          :disabled="selectedList.length === 0"
          @click="applyGlobalRole"
        >
          Apply
        </Button>
      </div>

      <div class="h-px bg-slate-100 my-4" />

      <div class="space-y-2">
        <div v-if="loading" class="text-sm text-gray-500">Loading users...</div>

        <div v-else>
          <div v-if="users.length === 0" class="text-sm text-gray-500">
            No users found
          </div>

          <div v-else class="space-y-2">
            <div
              v-for="u in users"
              :key="u.id"
              class="flex items-center justify-between px-3 py-2 border rounded"
            >
              <div class="flex items-center gap-3">
                <Checkbox
                  :model-value="isChecked(u)"
                  @update:model-value="
                    (v) => toggleUserSelection(u, v === true)
                  "
                />
                <div>
                  <div class="text-sm font-medium">{{ u.username }}</div>
                  <div class="text-xs text-gray-500">{{ u.email }}</div>
                </div>
              </div>

              <div class="flex items-center gap-2">
                <Select
                  :model-value="
                    isChecked(u)
                      ? selected[u.id]?.role
                      : (u.role ?? 'collaborator')
                  "
                  class="w-56"
                  :disabled="!isChecked(u)"
                  @update:model-value="
                    (val) => setUserRole(u.id, val as SelectedEntry['role'])
                  "
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="owner">Owner</SelectItem>
                    <SelectItem value="project_manager"
                      >Project Manager</SelectItem
                    >
                    <SelectItem value="collaborator">Collaborator</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="mt-4 flex items-center justify-between">
        <div class="text-sm text-gray-600">total {{ total }} users</div>

        <div class="flex items-center gap-2">
          <Button
            size="sm"
            variant="outline"
            :disabled="page <= 1"
            @click="goPage(page - 1)"
          >
            Prev
          </Button>

          <div class="flex items-center gap-1">
            <template v-for="p in totalPages" :key="p">
              <button
                class="px-2 py-1 rounded text-sm"
                :class="
                  p === page ? 'bg-blue-600 text-white' : 'hover:bg-slate-100'
                "
                @click="goPage(p)"
              >
                {{ p }}
              </button>
            </template>
          </div>

          <Button
            size="sm"
            variant="outline"
            :disabled="page >= totalPages"
            @click="goPage(page + 1)"
          >
            Next
          </Button>
        </div>
      </div>

      <DialogFooter class="mt-4">
        <div class="w-full flex justify-end gap-2">
          <Button variant="outline" @click="handleCancel">Cancel</Button>
          <Button :disabled="selectedList.length === 0" @click="handleConfirm">
            Add ({{ selectedList.length }})
          </Button>
        </div>
      </DialogFooter>
    </DialogContent>
  </Dialog>
  <SuccessDialog
    v-model:open="isSuccessDialogOpen"
    :message="successMessage"
    icon="fa fa-check-circle"
  />
</template>
