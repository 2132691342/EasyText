import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: [
      { find: '@', replacement: resolve(__dirname, 'src') },
      // wailsjs 在 frontend/ 根，Vite alias 把所有 ../wailsjs/ 解析到它
      { find: /^(\.\.\/)+wailsjs\/(.*)$/, replacement: resolve(__dirname, 'wailsjs') + '/$2' },
    ]
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
    sourcemap: false,
    minify: 'terser',
    rollupOptions: {
      output: {
        manualChunks: {
          // CodeMirror editor core
          codemirror: ['codemirror', '@codemirror/state', '@codemirror/view', '@codemirror/commands',
            '@codemirror/language', '@codemirror/autocomplete', '@codemirror/lint', '@codemirror/search'],
          // CodeMirror language packages (lazy-loaded per file type)
          cmLanguages: [
            '@codemirror/lang-json', '@codemirror/lang-javascript',
            '@codemirror/lang-html', '@codemirror/lang-css',
            '@codemirror/lang-markdown', '@codemirror/lang-xml',
            '@codemirror/lang-yaml', '@codemirror/lang-python',
            '@codemirror/lang-java', '@codemirror/lang-go',
            '@codemirror/lang-sql',
          ],
          // UI framework
          elementPlus: ['element-plus', '@element-plus/icons-vue'],
        }
      }
    }
  }
})