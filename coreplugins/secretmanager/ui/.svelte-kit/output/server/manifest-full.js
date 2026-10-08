export const manifest = (() => {
function __memo(fn) {
	let value;
	return () => value ??= (value = fn());
}

return {
	appDir: "_app",
	appPath: "secret-manager/pages/_app",
	assets: new Set(["icon/key.svg"]),
	mimeTypes: {".svg":"image/svg+xml"},
	_: {
		client: {start:"_app/immutable/entry/start.ko1_s6UR.js",app:"_app/immutable/entry/app.DN3K8vjK.js",imports:["_app/immutable/entry/start.ko1_s6UR.js","_app/immutable/chunks/FMI0s7r_.js","_app/immutable/chunks/DUcPvKhu.js","_app/immutable/chunks/C7JvW5Ae.js","_app/immutable/entry/app.DN3K8vjK.js","_app/immutable/chunks/DUcPvKhu.js","_app/immutable/chunks/CjusWKSe.js","_app/immutable/chunks/xxyg_ECW.js","_app/immutable/chunks/C7JvW5Ae.js","_app/immutable/chunks/kV5v4TTm.js","_app/immutable/chunks/DTrtZQqC.js"],stylesheets:[],fonts:[],uses_env_dynamic_public:false},
		nodes: [
			__memo(() => import('./nodes/0.js')),
			__memo(() => import('./nodes/1.js')),
			__memo(() => import('./nodes/2.js'))
		],
		remotes: {
			
		},
		routes: [
			{
				id: "/index.html",
				pattern: /^\/index\.html\/?$/,
				params: [],
				page: { layouts: [0,], errors: [1,], leaf: 2 },
				endpoint: null
			}
		],
		prerendered_routes: new Set([]),
		matchers: async () => {
			
			return {  };
		},
		server_assets: {}
	}
}
})();
