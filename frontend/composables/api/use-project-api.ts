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
  imageExtension?: string;
  imageData?: Stream;
}

export interface ProjectMemberUpdateReq {
  role: string;
}

export interface AddProjectMemberReq {
  userId: string;
  role: string;
}

export interface UserForAddMembers {
  role: null;
  id: string;
  username: string;
  email?: string | null;
  first_name?: string | null;
  last_name?: string | null;
}

export function getBaseProjectImageUrl(id: string) {
  const config = useRuntimeConfig();
  const base = getApiBaseUrlWithProtocol("http", config);
  const url = concatUrl(`assets/projects/${id}`, base);
  return url;
}

export function getUrlForProjectImage(id: string, imagePath: string) {
  const config = useRuntimeConfig();
  const extMatch = imagePath.match(/\.(\w+)$/);
  const ext = extMatch ? extMatch[1] : "png";
  const base = getApiBaseUrlWithProtocol("http", config);
  const encodedId = uuidToBase64Url(id);
  const url = concatUrl(`assets/projects/${encodedId}/file/${ext}`, base);
  url.searchParams.append("t", String(Date.now()));
  return url;
}

export function getProjectBaseUrl() {
  const config = useRuntimeConfig();
  const base = getApiBaseUrlWithProtocol("http", config);
  return concatUrl(`projects`, base);
}

export function useProject() {
  async function listProjects(page = 1, pageSize = 4) {
    const headers = useRequestHeaders(["cookie"]);
    return await $fetch<{ data: Project[]; count: number }>(
      getProjectBaseUrl().href,
      {
        method: "GET",
        query: { page: page, pageSize: pageSize },
        headers,
        credentials: "include",
      },
    );
  }

  async function createProject(formData: FormData) {
    return await $fetch<{ data: Project }>(getProjectBaseUrl().href, {
      method: "POST",
      body: formData,
      credentials: "include",
    });
  }

  async function updateProject(hexId: string, body: ProjectUpdateRequest) {
    const headers = useRequestHeaders(["cookie"]);
    const url = concatUrl(uuidToBase64Url(hexId), getProjectBaseUrl());
    return await $fetch<{ data: Project }>(url.href, {
      method: "PUT",
      body,
      headers,
      credentials: "include",
    });
  }

  async function deleteProject(hexId: string) {
    const url = concatUrl(uuidToBase64Url(hexId), getProjectBaseUrl());
    return await $fetch<{ data: Project }>(url.href, {
      method: "DELETE",
      credentials: "include",
    });
  }

  async function updateProjectImage(hexId: string, body: FormData) {
    const url = concatUrl(
      `${uuidToBase64Url(hexId)}/image`,
      getProjectBaseUrl(),
    );
    return await $fetch<{ imagePath: string; fileExtension: string }>(
      url.href,
      {
        method: "PUT",
        body,
        credentials: "include",
      },
    );
  }

  async function getProject(projectId: string) {
    const base = getProjectBaseUrl();
    const headers = useRequestHeaders(["cookie"]);

    return await $fetch<{ data: Project }>(
      concatUrl(projectId, base.href).href,
      {
        method: "GET",
        headers,
        credentials: "include",
      },
    );
  }

  async function getProjectMembers(projectId: string) {
    const headers = useRequestHeaders(["cookie"]);
    const base = getProjectBaseUrl();
    const url = concatUrl(`${projectId}/members`, base);
    return $fetch<{ data: ProjectMember[]; count: number }>(url.href, {
      method: "GET",
      headers,
      credentials: "include",
    });
  }

  async function getUsersForAddMembers(
    projectId: string,
    page = 1,
    pageSize = 10,
    search = "",
  ) {
    const headers = useRequestHeaders(["cookie"]);
    const base = getProjectBaseUrl();
    const url = concatUrl(`${projectId}/userForAddMembers`, base);
    return $fetch<{ data: UserForAddMembers[]; count: number }>(url.href, {
      method: "GET",
      query: { page, pageSize, search },
      headers,
      credentials: "include",
    });
  }

  async function addProjectMembers(
    projectId: string,
    body: AddProjectMemberReq[],
  ) {
    const base = getProjectBaseUrl();
    const url = concatUrl(`${projectId}/members`, base);
    return $fetch(url.href, {
      method: "POST",
      body,
      credentials: "include",
    });
  }

  async function removeProjectMember(projectId: string, memberId: string) {
    const base = getProjectBaseUrl();
    const url = concatUrl(`${projectId}/member/${memberId}`, base);
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
    const base = getProjectBaseUrl();
    const url = concatUrl(`${projectId}/user/${userId}/role`, base);
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
    getUsersForAddMembers,
    addProjectMembers,
    removeProjectMember,
    updateProjectMember,
  };
}
