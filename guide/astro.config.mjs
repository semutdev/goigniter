import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// https://astro.build/config
export default defineConfig({
  site: 'https://goigniter.semut.dev',
  integrations: [
    starlight({
      title: 'Goigniter',
      defaultLocale: 'en',
      locales: {
        en: {
          label: 'English',
          lang: 'en',
        },
        id: {
          label: 'Bahasa Indonesia',
          lang: 'id',
        },
      },
      favicon: '/favicon/favicon.ico',
      logo: {
        src: './src/assets/goigniter-logo.png',
        replacesTitle: false,
      },
      customCss: [
        './src/styles/custom.css',
      ],
      social: [
        {
          label: 'GitHub',
          href: 'https://github.com/semutdev/goigniter',
          icon: 'github',
        }
      ],
      sidebar: [
        {
          label: 'Guide',
          translations: {
            id: 'Panduan',
          },
          autogenerate: { directory: 'guide' },
        },
      ],
    }),
  ],
});