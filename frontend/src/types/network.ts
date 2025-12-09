export type RequestStatus = number;
export type ResourceType = 'document' | 'script' | 'stylesheet' | 'image' | 'font' | 'xhr' | 'other';

export interface NetworkRequest {
  id: string;
  url: string;
  method: string;
  status: RequestStatus;
  type: ResourceType;
  size: number | null;  // bytes
  time: number | null;  // seconds
  blocked?: boolean;
  failed?: boolean;
  isInternal: boolean;
}

export interface NetworkData {
  requests: NetworkRequest[];
}
