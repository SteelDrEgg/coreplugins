import adapter from '@sveltejs/adapter-static';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	kit: {
		adapter: adapter({
			pages: 'build',
			assets: 'build',
			fallback: undefined,
			precompress: false,
			strict: true
		}),
		alias: {
			$assets: 'src/assets',
			$components: 'src/components',
			$stores: 'src/stores',
			$utils: 'src/utils'
		},
		paths: {
			base: '/services/pages'
		},
		prerender: {
			handleHttpError: ({ path, message }) => {
				if (path === '/assets/css/scheme.css' || path === '/assets/js/sdk.js') return;
				throw new Error(message);
			}
		}
	}
};

export default config;
