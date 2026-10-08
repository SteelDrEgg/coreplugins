import * as universal from '../entries/pages/_layout.ts.js';

export const index = 0;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/_layout.svelte.js')).default;
export { universal };
export const universal_id = "src/routes/+layout.ts";
export const imports = ["_app/immutable/nodes/0.DGBCFCVe.js","_app/immutable/chunks/xxyg_ECW.js","_app/immutable/chunks/DUcPvKhu.js","_app/immutable/chunks/DTrtZQqC.js"];
export const stylesheets = ["_app/immutable/assets/0.Cz8Qd6Un.css"];
export const fonts = [];
