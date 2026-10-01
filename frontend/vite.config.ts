import { defineConfig } from 'vitest/config';
import { svelte } from '@sveltejs/vite-plugin-svelte';

// Alvo do proxy: o servidor Go do marco 5 (spec 05, decisão 10: sem CORS).
const api = process.env.BOMBER_API ?? 'http://localhost:8080';

export default defineConfig({
  plugins: [svelte()],
  server: {
    proxy: {
      '/partidas': api,
      '/bots': api,
      '/mapas': api,
    },
  },
  resolve: process.env.VITEST ? { conditions: ['browser'] } : undefined,
  test: {
    environment: 'jsdom',
    include: ['src/**/*.test.ts'],
  },
});
