import { uuidToBase64Url } from "~/lib/uuid";
import { getBaseProjectImageUrl, getProjectBaseUrl } from "./use-project-api";

export interface Model {
  modelId: string;
  projectId: string;
  name: string;
  description: string;
  version: number;
  imagePath?: string;
  imageExtension?: string;
  filePath?: string;
  modelExtension?: string;
  createdAt: string;
  updatedAt: string;
}
export interface ModelGetResponse {
  data: Model[];
  count: number;
}

export interface CreateModelResponse {
  data: Model;
}

export interface ModelCreateRequest {
  name: string;
  description: string;
  file: File;
}

export interface ModelUpdateRequest {
  name: string;
  description: string;
}

export function getUrlForModelImage(
  projectId: string,
  modelId: string,
  imageExt: string,
) {
  const ext = imageExt?.slice(1);
  const encodedModelId = uuidToBase64Url(modelId);
  const base = getBaseProjectImageUrl(projectId).href;
  const url = concatUrl(`models/${encodedModelId}/file/${ext}`, base);
  url.searchParams.append("t", String(Date.now()));
  return url;
}

export function getModelBaseUrl(id: string) {
  return concatUrl(`${id}/models`, getProjectBaseUrl());
}

export function useModels(projectId: string) {
  async function listModels(page = 1, pageSize = 4) {
    const headers = useRequestHeaders(["cookie"]);
    const base = getModelBaseUrl(projectId);
    return await $fetch<ModelGetResponse>(base.href, {
      method: "GET",
      query: {
        pageSize: pageSize,
        page: page,
      },
      headers,
      credentials: "include",
    });
  }

  async function postCreateModel(body: FormData) {
    return await $fetch<CreateModelResponse>(getModelBaseUrl(projectId).href, {
      method: "POST",
      body,
      credentials: "include",
    });
  }

  async function updateModel(modelId: string, body: ModelUpdateRequest) {
    const base = getModelBaseUrl(projectId);
    const url = concatUrl(modelId, base);
    return $fetch<CreateModelResponse>(url.href, {
      method: "PUT",
      body,
      credentials: "include",
    });
  }

  async function deleteModel(modelId: string) {
    const base = getModelBaseUrl(projectId);
    const url = concatUrl(modelId, base);
    return await $fetch(url.href, {
      method: "DELETE",
      credentials: "include",
    });
  }

  async function updateModelImage(modelId: string, body: FormData) {
    const base = getModelBaseUrl(projectId);
    const url = concatUrl(`${modelId}/image`, base);
    return await $fetch<{ imagePath: string; message: string }>(url.href, {
      method: "PUT",
      body,
      credentials: "include",
    });
  }

  return {
    listModels,
    postCreateModel,
    updateModel,
    deleteModel,
    updateModelImage,
  };
}
