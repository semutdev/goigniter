import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// https://astro.build/config
export default defineConfig({
  integrations: [
    starlight({
      title: 'Goigniter Docs',
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