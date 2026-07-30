import {
	getServiceDirectory,
	getServiceTempDirectory,
	listDiscoveredServices,
	listRunningServices,
	runServiceAction,
	updateServiceDirectory,
	updateServiceTempDirectory
} from '$lib/utils/api';
import * as m from '$lib/paraglide/messages.js';
import type {
	BannerMessage,
	DiscoveredService,
	MessageKind,
	RunningService,
	ServiceAction
} from '$lib/utils/types';

export class ServiceManager {
	discovered = $state<DiscoveredService[]>([]);
	running = $state<RunningService[]>([]);
	serviceDir = $state('');
	tempDir = $state('');
	tempDirRequiresRestart = $state(false);
	loaded = $state(false);
	message = $state<BannerMessage | null>(null);
	private busyCount = $state(0);

	get busy(): boolean {
		return this.busyCount > 0;
	}

	private showMessage(text: string, kind: MessageKind = 'info') {
		this.message = text ? { text, kind } : null;
	}

	private errorMessage(error: unknown, fallback: string): string {
		return error instanceof Error ? error.message : fallback;
	}

	private async withBusy<T>(operation: () => Promise<T>): Promise<T> {
		this.busyCount += 1;
		try {
			return await operation();
		} finally {
			this.busyCount -= 1;
		}
	}

	private async fetchState() {
		const [discovered, running, directory, tempDirectory] = await Promise.all([
			listDiscoveredServices(),
			listRunningServices(),
			getServiceDirectory(),
			getServiceTempDirectory()
		]);
		this.discovered = discovered;
		this.running = running;
		this.serviceDir = directory.service_dir || '';
		this.tempDir = tempDirectory.temp_dir || '';
		this.tempDirRequiresRestart = Boolean(tempDirectory.requires_restart);
	}

	async refresh(showNotice = false) {
		try {
			await this.withBusy(() => this.fetchState());
			if (showNotice) this.showMessage(m.service_list_refreshed(), 'success');
		} catch (error) {
			this.showMessage(this.errorMessage(error, m.failed_load_services()), 'error');
		} finally {
			this.loaded = true;
		}
	}

	async runAction(action: ServiceAction, name: string) {
		this.showMessage('');
		try {
			await this.withBusy(async () => {
				await runServiceAction(action, name);
				await this.fetchState();
			});
			const successMessage = {
				start: m.service_started,
				restart: m.service_restarted,
				stop: m.service_stopped
			}[action];
			this.showMessage(successMessage(), 'success');
		} catch (error) {
			const fallbackMessage = {
				start: m.failed_start_service,
				restart: m.failed_restart_service,
				stop: m.failed_stop_service
			}[action];
			this.showMessage(this.errorMessage(error, fallbackMessage()), 'error');
		}
	}

	async saveDirectories(serviceDir: string, tempDir: string) {
		if (serviceDir === this.serviceDir && tempDir === this.tempDir) {
			this.showMessage(m.no_directory_changes());
			return;
		}

		this.showMessage('');
		try {
			await this.withBusy(async () => {
				if (serviceDir !== this.serviceDir) await updateServiceDirectory(serviceDir);
				if (tempDir !== this.tempDir) await updateServiceTempDirectory(tempDir);
				await this.fetchState();
			});
			this.showMessage(
				this.tempDirRequiresRestart
					? m.directories_saved_restart()
					: m.directories_saved(),
				this.tempDirRequiresRestart ? 'warning' : 'success'
			);
		} catch (error) {
			this.showMessage(this.errorMessage(error, m.failed_save_directories()), 'error');
			await this.refresh();
		}
	}
}
