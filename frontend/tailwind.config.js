/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        bgMain: "var(--bg-main)",
        bgCard: "var(--bg-card)",
        bgPanel: "var(--bg-panel)",
        bgHover: "var(--bg-hover)",
        borderDark: "var(--border-color)",
        borderSubtle: "var(--border-subtle)",
        borderActive: "var(--border-active)",
        textMain: "var(--text-main)",
        textMuted: "var(--text-muted)",
        textFaint: "var(--text-faint)",
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', '-apple-system', 'sans-serif'],
        mono: ['"JetBrains Mono"', '"Fira Code"', 'monospace'],
      },
      boxShadow: {
        'card': '0 4px 20px -2px rgba(0, 0, 0, 0.5), 0 0 0 1px var(--border-color)',
      }
    },
  },
  plugins: [],
}
