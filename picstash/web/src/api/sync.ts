import { api } from './client'
import type { SyncLog, SyncFileLog, PaginatedResponse } from '@/types'

export const syncApi = {
  getLogs: (params: { page?: number; limit?: number }) =>
    api.get<PaginatedResponse<SyncLog>>('/sync/logs', { params }),

  getLogById: (id: number) =>
    api.get<{ log: SyncLog; file_logs: SyncFileLog[] }>(`/sync/logs/${id}`),

  getFileLogs: (id: number) =>
    api.get<{ data: SyncFileLog[] }>(`/sync/logs/${id}/files`),
}
