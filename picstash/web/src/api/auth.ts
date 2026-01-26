import { api } from './client'
import type { ApiResponse } from '@/types'

export const authApi = {
  sendCode: (email: string) =>
    api.post<ApiResponse<{ message: string; expires_in: number }>>('/auth/send-code', { email }),

  verifyCode: (email: string, code: string) =>
    api.post<ApiResponse<{ token: string; expires_at: string }>>('/auth/verify', { email, code }),
}
