import type { NetworkRequest, ResourceType } from '../types/network';

export interface ResourceTypeSummary {
  type: ResourceType;
  requests: number;
  totalTime: number;
  dataLoaded: number;
  failed: number;
}

export interface NetworkSummary {
  requests: number;
  internal: number;
  external: number;
  transferred: number;
  blocked: number;
  failed: number;
}

export function calculateNetworkSummary(requests: NetworkRequest[]): NetworkSummary {
  let internal = 0;
  let external = 0;
  let transferred = 0;
  let blocked = 0;
  let failed = 0;

  for (const request of requests) {
    if (request.isInternal) {
      internal++;
    } else {
      external++;
    }

    if (request.size !== null) {
      transferred += request.size;
    }

    if (request.blocked) {
      blocked++;
    }

    if (request.failed) {
      failed++;
    }
  }

  return {
    requests: requests.length,
    internal,
    external,
    transferred,
    blocked,
    failed,
  };
}

export function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';

  const units = ['B', 'KB', 'MB', 'GB'];
  const base = 1024;
  const unitIndex = Math.min(
    Math.floor(Math.log(bytes) / Math.log(base)),
    units.length - 1
  );

  const value = bytes / Math.pow(base, unitIndex);
  const formatted = unitIndex === 0 ? value.toString() : value.toFixed(1);

  return `${formatted} ${units[unitIndex]}`;
}

export function formatTimestamp(seconds: number): string {
  if (seconds < 1) {
    return `+${Math.round(seconds * 1000)}ms`;
  }
  return `+${seconds.toFixed(1)}s`;
}

export function formatTime(seconds: number): string {
  if (seconds === 0) return '0';
  if (seconds < 1) {
    return `${(seconds * 1000).toFixed(0)} ms`;
  }
  return `${seconds.toFixed(1)} sec`;
}

export function calculateResourceTypeSummary(requests: NetworkRequest[]): ResourceTypeSummary[] {
  const summaryMap = new Map<string, ResourceTypeSummary>();

  for (const request of requests) {
    const type = request.type || 'other';
    const existing = summaryMap.get(type);
    if (existing) {
      existing.requests++;
      existing.totalTime += request.time ?? 0;
      existing.dataLoaded += request.size ?? 0;
      if (request.failed) existing.failed++;
    } else {
      summaryMap.set(type, {
        type: type as ResourceType,
        requests: 1,
        totalTime: request.time ?? 0,
        dataLoaded: request.size ?? 0,
        failed: request.failed ? 1 : 0,
      });
    }
  }

  // Sort by data loaded descending
  return Array.from(summaryMap.values()).sort((a, b) => b.dataLoaded - a.dataLoaded);
}
