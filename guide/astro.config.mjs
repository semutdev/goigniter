import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// https://astro.build/config
export default defineConfig({
  integrations: [
    starlight({
      title: 'Goigniter',
      favicon: '/favicon/favicon.ico',
      logo: {
        src: './src/assets/goigniter-logo.png',
        replacesTitle: false, 
      },
      // accentColor: 'blue',
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
          label: 'User Guide',
          autogenerate: { directory: 'guide' }, 
        },
      ],
    }),
  ],
});