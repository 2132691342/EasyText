/// <reference types="vite/client" />

declare module '*.vue' {
    import type {DefineComponent} from 'vue'
    const component: DefineComponent<{}, {}, any>
    export default component
}

// Browser API augmentations
interface EyeDropperOpenOptions {
    signal?: AbortSignal
}
interface EyeDropperResult {
    sRGBHex: string
}
interface EyeDropperConstructor {
    new (): {
        open(options?: EyeDropperOpenOptions): Promise<EyeDropperResult>
    }
}
interface Window {
    EyeDropper?: EyeDropperConstructor
}
