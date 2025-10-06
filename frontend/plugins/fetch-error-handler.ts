import { useFailDialog } from "~/composables/useFailDialog";
import type { NitroFetchOptions } from "nitropack";

interface ErrorResponse {
  response?: {
    _data?: {
      message?: string;
    };
  };
  message?: string;
}

export default defineNuxtPlugin((_nuxtApp) => {
  const originalFetch = $fetch;

  const wrappedFetch = Object.assign(
    (async (
      request: string | Request,
      options: NitroFetchOptions<string | Request> = {},
    ) => {
      try {
        return await originalFetch(request, options);
      } catch (error: unknown) {
        const { showFailDialog } = useFailDialog();

        const message =
          (error as ErrorResponse)?.response?._data?.message ||
          (error as Error).message ||
          "Something went wrong. Please try again.";

        showFailDialog(message);

        throw error;
      }
    }) as typeof $fetch,
    {
      raw: originalFetch.raw?.bind(originalFetch),
      create: originalFetch.create?.bind(originalFetch),
    },
  ) as typeof $fetch;

  globalThis.$fetch = wrappedFetch;
});
