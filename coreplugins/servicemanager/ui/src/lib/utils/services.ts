import type { ConfigPatch, ServiceConfig, ServiceRow, DiscoveredService, RunningService, ServiceAction, ParamsPatch } from './types';

export function mergeServices(discovered: DiscoveredService[], running: RunningService[]): ServiceRow[] {
	const rows = new Map<string, ServiceRow>();
	for (const disk of discovered) {
		rows.set(disk.name, { name: disk.name, disk, instances: [], status: disk.status, versionMismatch: false });
	}
	for (const instance of running) {
		let row = rows.get(instance.name);
		if (!row) {
			row = { name: instance.name, instances: [], status: 'running', versionMismatch: false };
			rows.set(instance.name, row);
		}
		row.instances.push(instance);
		// The registry is authoritative for a live instance; preserve degraded/transition states.
		if (!['degraded', 'starting', 'stopping'].includes(row.status)) row.status = 'running';
		if (row.disk && row.disk.version !== instance.version) row.versionMismatch = true;
	}
	return [...rows.values()].sort((a, b) => a.name.localeCompare(b.name));
}

export function displayName(row: ServiceRow): string {
	const value = row.disk?.metadata?.DisplayName;
	return typeof value === 'string' && value.trim() ? value : row.name;
}

export function canAct(row: ServiceRow, action: ServiceAction): boolean {
	if (action === 'stop') {
		return row.instances.length > 0 || ['starting', 'stopping', 'backoff', 'failed'].includes(row.status);
	}
	if (!row.disk || ['starting', 'stopping'].includes(row.status)) return false;
	if (action === 'restart') return true;
	return row.instances.length === 0 && ['discovered', 'failed', 'exited'].includes(row.status);
}

// Only changed overrides are patched. Empty strings/null inherit; [] explicitly opens group access.
export function configPatch(before: ServiceConfig, after: ServiceConfig): ConfigPatch {
	const patch: ConfigPatch = {};
	for (const key of ['restart', 'run_as_user', 'checksum'] as const) {
		const value = after[key]?.trim() ?? '';
		if (value !== (before[key] ?? '')) patch[key] = value || null;
	}
	if (JSON.stringify(before.allow ?? null) !== JSON.stringify(after.allow ?? null)) {
		patch.allow = after.allow ?? null;
	}
	return patch;
}

export function paramsPatch(before: Record<string, string>, after: Record<string, string>): ParamsPatch {
	const set = Object.fromEntries(Object.entries(after).filter(([key, value]) => before[key] !== value));
	const remove = Object.keys(before).filter((key) => !Object.hasOwn(after, key));
	return { ...(Object.keys(set).length ? { set } : {}), ...(remove.length ? { remove } : {}) };
}
