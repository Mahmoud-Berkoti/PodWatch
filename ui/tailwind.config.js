/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: ['JetBrains Mono', 'Fira Code', 'monospace'],
        display: ['Space Grotesk', 'sans-serif'],
      },
      colors: {
        midnight: {
          900: '#0a0a0f',
          800: '#12121a',
          700: '#1a1a25',
          600: '#222230',
        },
        cyber: {
          green: '#ff6b00',
          blue: '#00d4ff',
          purple: '#b400ff',
          pink: '#ff0080',
          orange: '#ff6b00',
        },
        severity: {
          critical: '#ff2d55',
          high: '#ff9500',
          medium: '#ffcc00',
          low: '#34c759',
          info: '#5ac8fa',
        }
      },
      animation: {
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
        'glow': 'glow 2s ease-in-out infinite alternate',
      },
      keyframes: {
        glow: {
          '0%': { boxShadow: '0 0 5px rgba(255, 107, 0, 0.2)' },
          '100%': { boxShadow: '0 0 20px rgba(255, 107, 0, 0.6)' },
        }
      }
    },
  },
  plugins: [],
}
