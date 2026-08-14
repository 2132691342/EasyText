/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        primary: {
          50: '#eff6ff',
          100: '#dbeafe',
          200: '#bfdbfe',
          300: '#93c5fd',
          400: '#60a5fa',
          500: '#3b82f6',
          600: '#2563eb',
          700: '#1d4ed8',
          800: '#1e40af',
          900: '#1e3a8a',
        },
        editor: {
          bg: '#ffffff',
          lineHighlight: '#f5f5f5',
          selection: '#b3d7ff',
          cursor: '#000000',
          gutter: '#f0f0f0',
        },
        dark: {
          editor: {
            bg: '#1e1e1e',
            lineHighlight: '#2d2d2d',
            selection: '#264f78',
            cursor: '#ffffff',
            gutter: '#252526',
          }
        }
      },
      fontFamily: {
        mono: ['Consolas', 'Monaco', 'Courier New', 'monospace'],
      },
      fontSize: {
        'editor': '14px',
      },
      lineHeight: {
        'editor': '1.5',
      },
    },
  },
  plugins: [],
}