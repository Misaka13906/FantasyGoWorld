import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    environment: 'jsdom',
    globals: true,
    include: ['tests/components/**/*.test.ts', 'tests/components/**/*.test.tsx'],
  },
});
