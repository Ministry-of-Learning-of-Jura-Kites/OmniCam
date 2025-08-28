export interface Model {
  id: string;
  projectId: string;
  name: string;
  description: string;
  version: number;
  createdAt: string;
  updatedAt: string;
}

export interface ModelCreateRequest {
  name: string;
  description: string;
}

export interface ModelUpdateRequest {
  name: string;
  description: string;
}
