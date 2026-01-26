import { defaultWatermarkConfig } from '@/types'
import type { WatermarkConfig } from '@/types'

const STORAGE_KEY = 'watermark_config'

export function saveWatermarkConfig(config: WatermarkConfig): void {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(config))
  } catch (e) {
    console.error('Failed to save watermark config:', e)
  }
}

export function loadWatermarkConfig(): WatermarkConfig {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored) {
      const parsed = JSON.parse(stored)
      return { ...defaultWatermarkConfig, ...parsed }
    }
  } catch (e) {
    console.error('Failed to load watermark config:', e)
  }
  return { ...defaultWatermarkConfig }
}

export function clearWatermarkConfig(): void {
  try {
    localStorage.removeItem(STORAGE_KEY)
  } catch (e) {
    console.error('Failed to clear watermark config:', e)
  }
}
