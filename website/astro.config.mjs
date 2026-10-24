// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

export default defineConfig({
	site: 'https://vardius.github.io',
	base: '/message-bus',
	trailingSlash: 'always',
	// Pages import the runnable programs in ../examples, so the dev server may read outside this folder.
	vite: { server: { fs: { allow: ['..'] } } },
	integrations: [
		starlight({
			title: 'message-bus',
			description:
				'A typed message bus for Go: publish to topics, and every subscriber handles the message in order on its own goroutine.',
			logo: { src: './src/assets/mark.svg', alt: '' },
			favicon: '/favicon.svg',
			components: { Hero: './src/components/Hero.astro' },
			// Long lines wrap instead of scrolling under the copy button on phones.
			expressiveCode: { defaultProps: { wrap: true } },
			social: [{ icon: 'github', label: 'GitHub', href: 'https://github.com/vardius/message-bus' }],
			customCss: ['./src/styles/custom.css'],
			sidebar: [
				{
					label: 'Start here',
					items: ['getting-started'],
				},
				{
					label: 'The bus',
					items: ['publishing', 'subscribing', 'closing'],
				},
				{
					label: 'Reference',
					items: [
						'benchmarks',
						'migrating-from-v1',
						{ label: 'API reference', link: 'https://pkg.go.dev/github.com/vardius/message-bus/v2' },
					],
				},
			],
		}),
	],
});
