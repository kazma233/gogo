/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_DOMAIN: string | undefined
  readonly VITE_BASE_URL: string | undefined
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
