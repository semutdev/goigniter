import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// https://astro.build/config
export default defineConfig({
  site: 'https://goigniter.semut.dev',
  redirects: {
    '/': '/en',
    '/guide': '/en/guide',
    '/guide/01-intro': '/en/guide/01-intro',
    '/guide/02-installation': '/en/guide/02-installation',
    '/guide/03-routing': '/en/guide/03-routing',
    '/guide/04-controllers': '/en/guide/04-controllers',
    '/guide/05-middleware': '/en/guide/05-middleware',
    '/guide/06-templates': '/en/guide/06-templates',
    '/guide/07-database': '/en/guide/07-database',
    '/guide/08-query-builder': '/en/guide/08-query-builder',
    '/guide/09-helpers': '/en/guide/09-helpers',
    '/guide/10-upload': '/en/guide/10-upload',
    '/guide/11-agentic': '/en/guide/11-agentic',
  },
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
          items: [
            'guide/01-intro',
            'guide/02-installation',
            'guide/03-routing',
            'guide/04-controllers',
            'guide/05-middleware',
            'guide/06-templates',
            'guide/07-database',
            'guide/08-query-builder',
            'guide/09-helpers',
            'guide/10-upload',
            'guide/11-agentic',
          ],
        },
      ],
    }),
  ],
});