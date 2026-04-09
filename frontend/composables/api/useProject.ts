import type Stream from "stream";
import { uuidToBase64Url } from "~/lib/uuid";
import { getApiBaseUrlWithProtocol } from "~/utils/url";

export interface ProjectUpdateRequest {
  name: string;
  description: string;
}

export interface ProjectMember {
  userId: string;
  username: string;
  firstName?: string | null;
  lastName?: string | null;
  role: string;
}

export interface Project {
  id: string;
  name: string;
  description: string;
  createdAt: string;
  updatedAt: string;
  imagePath?: string;
  imageData?: Stream;
}

export interface ProjectMemberUpdateReq {
  role: string;
}

export function getBaseProjectImageUrl(id: string) {
  const config = useRuntimeConfig();
  const base = getApiBaseUrlWithProtocol("http", config, false);
  const url = new URL(`assets/projects/${id}`, base);
  return url;
}

export function getUrlForProjectImage(id: string, imagePath: string) {
  const config = useRuntimeConfig();
  const extMatch = imagePath.match(/\.(\w+)$/);
  const ext = extMatch ? extMatch[1] : "png";
  const base = getApiBaseUrlWithProtocol("http", config, false);
  const url = new URL(`assets/projects/${id}/file/${ext}`, base);
  const now = Date.now();
  url.searchParams.append("t", String(now));
  return url;
}

export function getProjectBaseUrl(allowServerSide: boolean) {
  const config = useRuntimeConfig();
  const base = getApiBaseUrlWithProtocol("http", config, allowServerSide);
  return new URL(`projects`, base);
}

export function useProject() {
  async function listProjects(page = 1, pageSize = 4) {
    const response = await useFetch<{ data: Project[]; count: number }>(
      getProjectBaseUrl(true).href,
      {
        method: "GET",
        query: { page: page, pageSize: pageSize },
        credentials: "include",
      },
    );
    return response;
  }

  async function createProject(formData: FormData) {
    return await $fetch<{ data: Project }>(getProjectBaseUrl(false).href, {
      method: "POST",
      body: formData,
      credentials: "include",
    });
  }

  async function updateProject(hexId: string, body: ProjectUpdateRequest) {
    const url = new URL(uuidToBase64Url(hexId), getProjectBaseUrl(true));
    return await $fetch<{ data: Project }>(url.href, {
      method: "PUT",
      body,
      credentials: "include",
    });
  }

  async function deleteProject(hexId: string) {
    const url = new URL(uuidToBase64Url(hexId), getProjectBaseUrl(true));
    return await $fetch<{ data: Project }>(url.href, {
      method: "DELETE",
      credentials: "include",
    });
  }

  async function updateProjectImage(hexId: string, body: FormData) {
    const url = new URL(
      `${uuidToBase64Url(hexId)}/image`,
      addTrailingSlash(getProjectBaseUrl(true)),
    );
    return await $fetch<{ imagePath: string }>(url.href, {
      method: "PUT",
      body,
      credentials: "include",
    });
  }

  async function getProject(projectId: string) {
    const base = getProjectBaseUrl(true);
    return await useFetch<{ data: Project }>(
      new URL(projectId, addTrailingSlash(base.href)).href,
      {
        method: "GET",
        credentials: "include",
      },
    );
  }

  async function getProjectMembers(projectId: string) {
    const base = getProjectBaseUrl(true);
    const url = new URL(`${projectId}/members`, addTrailingSlash(base));
    return $fetch<{ data: ProjectMember[]; count: number }>(url.href, {
      method: "GET",
      credentials: "include",
    });
  }

  async function removeProjectMember(projectId: string, memberId: string) {
    const base = getProjectBaseUrl(true);
    const url = new URL(
      `${projectId}/members/${memberId}`,
      addTrailingSlash(base),
    );
    return $fetch<{ data: ProjectMember[]; count: number }>(url.href, {
      method: "DELETE",
      credentials: "include",
    });
  }

  async function updateProjectMember(
    projectId: string,
    userId: string,
    body: ProjectMemberUpdateReq,
  ) {
    const base = getProjectBaseUrl(false);
    const url = new URL(
      `${projectId}/user/${userId}/role`,
      addTrailingSlash(base),
    );
    return $fetch(url.href, {
      method: "PUT",
      body,
      credentials: "include",
    });
  }

  return {
    listProjects,
    createProject,
    updateProject,
    updateProjectImage,
    deleteProject,
    getProject,
    getProjectMembers,
    removeProjectMember,
    updateProjectMember,
  };
}
