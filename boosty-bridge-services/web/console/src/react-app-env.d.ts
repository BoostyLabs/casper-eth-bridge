/// <reference types="react-scripts" />

declare module '*.svg' {
    const src: string;
    export default src;
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
