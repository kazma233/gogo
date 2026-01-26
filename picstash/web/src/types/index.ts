export interface WatermarkConfig {
  enabled: boolean
  text: string
  position: 'top-left' | 'top-center' | 'top-right' | 'center-left' | 'center' | 'center-right' | 'bottom-left' | 'bottom-center' | 'bottom-right'
  size: number
  color: string
  opacity: number
  fullscreen: boolean
  spacing: number
  rotation: number
  padding: number
}

export const defaultWatermarkConfig: WatermarkConfig = {
  enabled: false,
  text: '',
  position: 'bottom-right',
  size: 24,
  color: '#FFFFFF',
  opacity: 0.7,
  fullscreen: false,
  spacing: 20,
  rotation: 0,
  padding: 0,
}

export interface Image {
  id: number
  path: string
  url: string
  sha?: string
  thumbnail_path?: string
  thumbnail_url?: string
  thumbnail_sha?: string
  thumbnail_size?: number
  watermark_path?: string
  watermark_url?: string
  watermark_sha?: string
  watermark_size?: number
  original_filename: string
  filename: string
  size?: number
  mime_type: string
  width?: number
  height?: number
  has_thumbnail: boolean
  has_watermark: boolean
  uploaded_at: string
  tags?: Tag[]
}

export interface Tag {
  id: number
  name: string
  color: string
  created_at: string
}

export interface User {
  email: string
  token?: string
}

export interface ApiResponse<T = any> {
  data?: T
  total?: number
  page?: number
  limit?: number
  error?: string
  message?: string
}

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  limit: number
}

export interface SyncLog {
  id: number
  triggered_by: string
  started_at: string
  completed_at: string | null
  status: 'running' | 'completed' | 'completed_with_errors' | 'failed'
  total_files: number
  processed_files: number
  error_count: number
  error_message: string | null
}

export interface SyncFileLog {
  id: number
  sync_log_id: number
  path: string
  action: 'created' | 'updated' | 'deleted' | 'skipped'
  status: 'success' | 'failed'
  sha: string | null
  old_sha: string | null
  size: number | null
  old_size: number | null
  error_message: string | null
  created_at: string
}

export interface SyncResult {
  created_count: number
  updated_count: number
  deleted_count: number
  skipped_count: number
  error_count: number
  log_id: number
}
