// ~/composables/useWorkspaceApi.ts
import type { RuntimeConfig } from "nuxt/schema";
import { MODEL_INFO_KEY } from "~/constants/state-keys";
import type { Simulation } from "~/types/simulation";

export function useWorkspaceApi(
  projectId: string,
  modelId: string,
  workspaceId: string,
  runtimeConfig: RuntimeConfig,
) {
  const route = useWorkspaceRoute(
    projectId,
    modelId,
    workspaceId,
    runtimeConfig,
  );

  async function postMerge() {
    const resp = await fetch(route.merge(), {
      method: "POST",
      credentials: "include",
    });

    return { resp };
  }

  async function postCreateMe() {
    const data = await $fetch(route.me(), {
      method: "POST",
      credentials: "include",
    });

    useState(`${MODEL_INFO_KEY}-${modelId}`, () => data);
  }

  async function deleteWorkspaceMe() {
    await $fetch(route.me(), {
      method: "DELETE",
      credentials: "include",
    });

    useState(`${MODEL_INFO_KEY}-${modelId}`, () => undefined);
  }

  async function postResolve(results: Record<string, Record<string, unknown>>) {
    const error = ref<Error | undefined>();

    try {
      await $fetch(route.resolve(), {
        method: "POST",
        credentials: "include",
        body: { merged: results },
      });
    } catch (err) {
      error.value = err as Error;
      console.error("postResolve failed:", err);
    }

    return { error };
  }

  async function putSimulation(simulation: Simulation) {
    console.log("update simulation:", simulation);

    return await $fetch(route.simulation(), {
      method: "PUT",
      credentials: "include",
      body: { simulation },
    });
  }

  async function getWorkspace(fields: string[] = []) {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const data = await $fetch<{ data: any }>(route.workspaceIdUrl(), {
      method: "GET",
      credentials: "include",
      query: {
        fields,
      },
    });

    return data.data;
  }

  return {
    getWorkspace,
    postMerge,
    postCreateMe,
    deleteWorkspaceMe,
    postResolve,
    putSimulation,
  };
}
