import { api } from './client'
import type { ApiResponse, Image, SyncResult } from '@/types'

export const imagesApi = {
  getList: (params: { page?: number; limit?: number; tag_id?: number }) =>
    api.get<{ data: Image[]; total: number; page: number; limit: number }>('/images', { params }),

  getById: (id: number) =>
    api.get<ApiResponse<Image>>(`/images/${id}`),

  upload: (formData: FormData) =>
    api.post<ApiResponse<{ data: Image[]; count: number }>>('/images/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      timeout: 120000,
    }),

  delete: (id: number) =>
    api.delete<ApiResponse>(`/images/${id}`),

  updateTags: (id: number, tagIds: number[]) =>
    api.put<ApiResponse>(`/images/${id}/tags`, { tag_ids: tagIds }),

  sync: () =>
    api.post<ApiResponse<SyncResult>>('/images/sync'),
}
