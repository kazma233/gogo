import type { WatermarkConfig } from '@/types'

export function createWatermarkCanvas(
  image: HTMLImageElement,
  config: WatermarkConfig
): HTMLCanvasElement {
  const canvas = document.createElement('canvas')
  canvas.width = image.naturalWidth
  canvas.height = image.naturalHeight
  const ctx = canvas.getContext('2d')!

  ctx.drawImage(image, 0, 0)

  if (!config.enabled || !config.text) {
    return canvas
  }

  const { text, position, size, color, opacity, fullscreen, spacing, rotation } = config

  ctx.save()
  ctx.globalAlpha = opacity
  ctx.font = `${size}px sans-serif`
  ctx.fillStyle = color
  ctx.textBaseline = 'top'

  const textWidth = ctx.measureText(text).width
  const textHeight = size

  const rotationRad = (rotation * Math.PI) / 180
  const cos = Math.abs(Math.cos(rotationRad))
  const sin = Math.abs(Math.sin(rotationRad))
  const rotatedWidth = textWidth * cos + textHeight * sin
  const rotatedHeight = textWidth * sin + textHeight * cos

  const padding = config.padding || 0
  const canvasWidth = ctx.canvas.width
  const canvasHeight = ctx.canvas.height

  if (fullscreen) {
    drawFullscreenWatermark(ctx, text, spacing, rotation, rotatedWidth, rotatedHeight)
  } else {
    let drawX = 0
    let drawY = 0

    switch (position) {
      case 'top-left':
        drawX = padding
        drawY = padding
        break
      case 'top-center':
        drawX = (canvasWidth - rotatedWidth) / 2
        drawY = padding
        break
      case 'top-right':
        drawX = canvasWidth - rotatedWidth - padding
        drawY = padding
        break
      case 'center-left':
        drawX = padding
        drawY = (canvasHeight - rotatedHeight) / 2
        break
      case 'center':
        drawX = (canvasWidth - rotatedWidth) / 2
        drawY = (canvasHeight - rotatedHeight) / 2
        break
      case 'center-right':
        drawX = canvasWidth - rotatedWidth - padding
        drawY = (canvasHeight - rotatedHeight) / 2
        break
      case 'bottom-left':
        drawX = padding
        drawY = canvasHeight - rotatedHeight - padding
        break
      case 'bottom-center':
        drawX = (canvasWidth - rotatedWidth) / 2
        drawY = canvasHeight - rotatedHeight - padding
        break
      case 'bottom-right':
        drawX = canvasWidth - rotatedWidth - padding
        drawY = canvasHeight - rotatedHeight - padding
        break
    }

    drawRotatedText(ctx, text, drawX, drawY, rotation, rotatedWidth, rotatedHeight)
  }

  ctx.restore()

  return canvas
}

function drawFullscreenWatermark(
  ctx: CanvasRenderingContext2D,
  text: string,
  spacing: number,
  rotation: number,
  rotatedWidth: number,
  rotatedHeight: number
) {
  const canvasWidth = ctx.canvas.width
  const canvasHeight = ctx.canvas.height

  const cols = Math.ceil(canvasWidth / (rotatedWidth + spacing))
  const rows = Math.ceil(canvasHeight / (rotatedHeight + spacing))

  for (let row = -1; row < rows + 1; row++) {
    for (let col = -1; col < cols + 1; col++) {
      const x = col * (rotatedWidth + spacing)
      const y = row * (rotatedHeight + spacing)
      drawRotatedText(ctx, text, x, y, rotation, rotatedWidth, rotatedHeight)
    }
  }
}

function drawRotatedText(
  ctx: CanvasRenderingContext2D,
  text: string,
  x: number,
  y: number,
  rotation: number,
  rotatedWidth: number,
  rotatedHeight: number
) {
  const rotationRad = (rotation * Math.PI) / 180

  ctx.save()
  ctx.translate(x + rotatedWidth / 2, y + rotatedHeight / 2)
  ctx.rotate(rotationRad)
  ctx.textBaseline = 'middle'
  ctx.textAlign = 'center'
  ctx.fillText(text, 0, 0)
  ctx.restore()
}

export function canvasToBlob(canvas: HTMLCanvasElement, quality: number = 0.9): Promise<Blob> {
  return new Promise((resolve, reject) => {
    canvas.toBlob(
      (blob) => {
        if (blob) {
          resolve(blob)
        } else {
          reject(new Error('Failed to convert canvas to blob'))
        }
      },
      'image/jpeg',
      quality
    )
  })
}

export function getImageDimensions(file: File): Promise<{ width: number; height: number }> {
  return new Promise((resolve, reject) => {
    const img = new Image()
    img.onload = () => {
      resolve({ width: img.naturalWidth, height: img.naturalHeight })
      URL.revokeObjectURL(img.src)
    }
    img.onerror = reject
    img.src = URL.createObjectURL(file)
  })
}
