<script lang="ts">
	import { api } from '$lib/api';
	import { onMount } from 'svelte';
	import ConfirmModal from '$lib/components/ConfirmModal.svelte';
	import type { Webhook, WebhookDelivery, CreateWebhookRequest, UpdateWebhookRequest } from '$lib/types';
	import { WEBHOOK_EVENTS, type WebhookEventType } from '$lib/types';

	// Tab management
	let activeTab = $state('general');
	const tabs = [
		{ id: 'general', label: 'عمومی' },
		{ id: 'video', label: 'ویدیو' },
		{ id: 'security', label: 'امنیت' },
		{ id: 'email', label: 'ایمیل' },
		{ id: 'api', label: 'API' },
		{ id: 'webhooks', label: 'وب‌هوک‌ها' },
	];

	// General settings state
	let settings = $state({
		max_users_per_room: 100,
		recording_enabled: true,
		maintenance_mode: false,
		allow_student_video: false,
		max_file_size_mb: 50,
		session_auto_end_minutes: 120,
		// Security
		password_min_length: '6',
		password_require_uppercase: false,
		password_require_number: false,
		password_require_special: false,
		session_timeout_minutes: '60',
		max_login_attempts: '5',
		lockout_duration_minutes: '30',
		require_2fa: false,
		// Email
		smtp_enabled: false,
		smtp_host: 'smtp.gmail.com',
		smtp_port: '587',
		smtp_username: '',
		smtp_password: '',
		smtp_from: 'noreply@iroom.ir',
		// API
		external_api_key: '',
	});
	let loading = $state(true);
	let saving = $state(false);
	let saved = $state(false);

	// Email test state
	let emailTesting = $state(false);
	let emailTestResult = $state<'success' | 'error' | null>(null);

	// Webhook state
	let webhooks = $state<Webhook[]>([]);
	let webhooksLoading = $state(true);
	let showWebhookModal = $state(false);
	let editingWebhook = $state<Webhook | null>(null);
	let webhookSaving = $state(false);
	let webhookTested = $state<number | null>(null);

	// Webhook form state
	let webhookForm = $state<CreateWebhookRequest>({
		url: '',
		events: [],
	});
	let webhookActive = $state(true);

	// Delivery logs state
	let deliveries = $state<WebhookDelivery[]>([]);
	let deliveriesLoading = $state(false);
	let selectedWebhookId = $state<number | null>(null);
	let showDeliveries = $state(false);

	// Delete confirmation
	let showDeleteConfirm = $state(false);
	let webhookToDelete = $state<number | null>(null);

	// Password change state
	let currentPassword = $state('');
	let newPassword = $state('');
	let confirmPassword = $state('');
	let passwordChangeLoading = $state(false);
	let passwordChangeSuccess = $state(false);
	let passwordChangeError = $state('');

	async function changeMyPassword() {
		if (newPassword !== confirmPassword) {
			passwordChangeError = 'رمز عبور جدید و تکرار آن مطابقت ندارند';
			return;
		}
		if (newPassword.length < 6) {
			passwordChangeError = 'رمز عبور باید حداقل ۶ کاراکتر باشد';
			return;
		}
		passwordChangeLoading = true;
		passwordChangeError = '';
		passwordChangeSuccess = false;
		const res = await api.post('/auth/change-password', {
			old_password: currentPassword,
			new_password: newPassword,
		});
		if (res.success) {
			passwordChangeSuccess = true;
			currentPassword = '';
			newPassword = '';
			confirmPassword = '';
			setTimeout(() => passwordChangeSuccess = false, 3000);
		} else {
			passwordChangeError = res.error || 'خطا در تغییر رمز عبور';
		}
		passwordChangeLoading = false;
	}

	onMount(async () => {
		const res = await api.get<any>('/admin/settings');
		if (res.success && res.data) settings = { ...settings, ...res.data };
		loading = false;

		// Load webhooks
		await loadWebhooks();
	});

	async function saveSettings() {
		saving = true;
		saved = false;
		const res = await api.put('/admin/settings', settings);
		if (res.success) {
			saved = true;
			setTimeout(() => saved = false, 3000);
		}
		saving = false;
	}

	// Webhook functions
	async function loadWebhooks() {
		webhooksLoading = true;
		const res = await api.get<Webhook[]>('/admin/webhooks');
		if (res.success && res.data) {
			webhooks = res.data;
		}
		webhooksLoading = false;
	}

	function openCreateModal() {
		editingWebhook = null;
		webhookForm = { url: '', events: [] };
		webhookActive = true;
		showWebhookModal = true;
	}

	function openEditModal(webhook: Webhook) {
		editingWebhook = webhook;
		webhookForm = { url: webhook.url, events: [...webhook.events] };
		webhookActive = webhook.is_active;
		showWebhookModal = true;
	}

	function closeWebhookModal() {
		showWebhookModal = false;
		editingWebhook = null;
		webhookForm = { url: '', events: [] };
		webhookActive = true;
	}

	function toggleEvent(event: WebhookEventType) {
		if (webhookForm.events.includes(event)) {
			webhookForm.events = webhookForm.events.filter(e => e !== event);
		} else {
			webhookForm.events = [...webhookForm.events, event];
		}
	}

	async function saveWebhook() {
		if (!webhookForm.url || webhookForm.events.length === 0) return;

		webhookSaving = true;
		let res;
		if (editingWebhook) {
			const updateData: UpdateWebhookRequest = {
				url: webhookForm.url,
				events: webhookForm.events,
				is_active: webhookActive,
			};
			res = await api.put<Webhook>(`/admin/webhooks/${editingWebhook.id}`, updateData);
		} else {
			res = await api.post<Webhook>('/admin/webhooks', webhookForm);
		}

		if (res.success) {
			await loadWebhooks();
			closeWebhookModal();
		}
		webhookSaving = false;
	}

	function confirmDeleteWebhook(id: number) {
		webhookToDelete = id;
		showDeleteConfirm = true;
	}

	async function deleteWebhook() {
		if (!webhookToDelete) return;
		const res = await api.delete(`/admin/webhooks/${webhookToDelete}`);
		if (res.success) {
			await loadWebhooks();
			if (selectedWebhookId === webhookToDelete) {
				showDeliveries = false;
				selectedWebhookId = null;
			}
		}
		showDeleteConfirm = false;
		webhookToDelete = null;
	}

	async function testWebhook(id: number) {
		webhookTested = id;
		const res = await api.post(`/admin/webhooks/${id}/test`);
		if (res.success) {
			setTimeout(() => webhookTested = null, 3000);
		}
	}

	async function loadDeliveries(webhookId: number) {
		if (selectedWebhookId === webhookId && showDeliveries) {
			showDeliveries = false;
			selectedWebhookId = null;
			return;
		}

		selectedWebhookId = webhookId;
		showDeliveries = true;
		deliveriesLoading = true;
		const res = await api.get<{ items: WebhookDelivery[] }>(`/admin/webhooks/${webhookId}/deliveries`);
		if (res.success && res.data) {
			deliveries = res.data.items;
		}
		deliveriesLoading = false;
	}

	function getEventLabel(event: string): string {
		return WEBHOOK_EVENTS[event as WebhookEventType] || event;
	}

	function formatDate(dateStr: string): string {
		return new Date(dateStr).toLocaleString('fa-IR');
	}

	// Email test
	async function testEmail() {
		emailTesting = true;
		emailTestResult = null;
		const res = await api.post('/admin/settings/test-email', {
			smtp_host: settings.smtp_host,
			smtp_port: parseInt(settings.smtp_port as string) || 587,
			smtp_username: settings.smtp_username,
			smtp_password: settings.smtp_password,
			smtp_from: settings.smtp_from,
		});
		emailTestResult = res.success ? 'success' : 'error';
		emailTesting = false;
		setTimeout(() => emailTestResult = null, 5000);
	}
</script>

<div style="max-width:800px;margin:0 auto;" class="space-y-5">
	<div>
		<h1 class="sky-page-title">تنظیمات سیستم</h1>
		<p class="sky-page-subtitle">تنظیمات کلی پلتفرم کلاس آنلاین</p>
	</div>

	<!-- Tabs -->
	<div class="sky-card overflow-hidden">
		<div class="sky-tabs" style="padding: 0 1.25rem; overflow-x: auto;">
			{#each tabs as tab}
				<button class="sky-tab {activeTab === tab.id ? 'active' : ''}" onclick={() => activeTab = tab.id}>
					{tab.label}
				</button>
			{/each}
		</div>

		<div class="p-6">
			{#if activeTab === 'general'}
				<!-- General Settings -->
				{#if loading}
					<div class="flex items-center justify-center py-12">
						<svg class="sky-spinner lg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--color-crystal-clear);"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/></svg>
					</div>
				{:else}
					<div class="divide-y">
						<!-- Max users per room -->
						<div class="py-4 flex items-center justify-between">
							<div>
								<p class="font-medium" style="color: var(--color-midnight-sky);">حداکثر کاربر در اتاق</p>
								<p class="text-sm mt-0.5" style="color: var(--color-mystic-sea);">تعداد حداکثر شرکت‌کنندگان در هر جلسه</p>
							</div>
							<input type="number" bind:value={settings.max_users_per_room} min="2" max="500" class="sky-input text-center" style="width: 5rem;" />
						</div>

						<!-- Recording enabled -->
						<div class="py-4 flex items-center justify-between">
							<div>
								<p class="font-medium" style="color: var(--color-midnight-sky);">ضبط جلسات</p>
								<p class="text-sm mt-0.5" style="color: var(--color-mystic-sea);">امکان ضبط جلسات توسط مدرس</p>
							</div>
							<button
								onclick={() => settings.recording_enabled = !settings.recording_enabled}
								class="relative w-11 h-6 rounded-full transition-colors {settings.recording_enabled ? 'bg-[var(--color-crystal-clear)]' : 'bg-[var(--color-muted-mountain)]'}"
							>
								<span class="absolute top-0.5 right-0.5 w-5 h-5 bg-white rounded-full transition-transform {settings.recording_enabled ? 'translate-x-[-20px]' : ''}"></span>
							</button>
						</div>

						<!-- Allow student video -->
						<div class="py-4 flex items-center justify-between">
							<div>
								<p class="font-medium" style="color: var(--color-midnight-sky);">ارسال ویدیو توسط دانش‌آموز</p>
								<p class="text-sm mt-0.5" style="color: var(--color-mystic-sea);">اجازه ارسال ویدیو به دانش‌آموزان</p>
							</div>
							<button
								onclick={() => settings.allow_student_video = !settings.allow_student_video}
								class="relative w-11 h-6 rounded-full transition-colors {settings.allow_student_video ? 'bg-[var(--color-crystal-clear)]' : 'bg-[var(--color-muted-mountain)]'}"
							>
								<span class="absolute top-0.5 right-0.5 w-5 h-5 bg-white rounded-full transition-transform {settings.allow_student_video ? 'translate-x-[-20px]' : ''}"></span>
							</button>
						</div>

						<!-- Max file size -->
						<div class="py-4 flex items-center justify-between">
							<div>
								<p class="font-medium" style="color: var(--color-midnight-sky);">حداکثر حجم فایل (MB)</p>
								<p class="text-sm mt-0.5" style="color: var(--color-mystic-sea);">حداکثر اندازه آپلود فایل</p>
							</div>
							<input type="number" bind:value={settings.max_file_size_mb} min="1" max="500" class="sky-input text-center" style="width: 5rem;" />
						</div>

						<!-- Auto end session -->
						<div class="py-4 flex items-center justify-between">
							<div>
								<p class="font-medium" style="color: var(--color-midnight-sky);">پایان خودکار جلسه (دقیقه)</p>
								<p class="text-sm mt-0.5" style="color: var(--color-mystic-sea);">زمان پایان خودکار پس از شروع</p>
							</div>
							<input type="number" bind:value={settings.session_auto_end_minutes} min="30" max="480" class="sky-input text-center" style="width: 5rem;" />
						</div>

						<!-- Maintenance mode -->
						<div class="py-4 flex items-center justify-between">
							<div>
								<p class="font-medium" style="color: var(--color-midnight-sky);">حالت تعمیر و نگهداری</p>
								<p class="text-sm mt-0.5" style="color: var(--color-mystic-sea);">غیرفعال‌سازی موقت سیستم</p>
							</div>
							<button
								onclick={() => settings.maintenance_mode = !settings.maintenance_mode}
								class="relative w-11 h-6 rounded-full transition-colors {settings.maintenance_mode ? 'bg-[var(--color-fiery-passion)]' : 'bg-[var(--color-muted-mountain)]'}"
							>
								<span class="absolute top-0.5 right-0.5 w-5 h-5 bg-white rounded-full transition-transform {settings.maintenance_mode ? 'translate-x-[-20px]' : ''}"></span>
							</button>
						</div>
					</div>

					<div class="flex items-center justify-between mt-6 pt-4 border-t">
						{#if saved}
							<span class="text-sm" style="color: var(--color-lush-meadow);">ذخیره شد</span>
						{:else}
							<span></span>
						{/if}
						<button onclick={saveSettings} disabled={saving} class="px-6 py-2.5 bg-blue-600 text-white rounded-lg font-medium text-sm hover:bg-blue-700 transition-colors disabled:opacity-50">
							{saving ? 'در حال ذخیره...' : 'ذخیره تنظیمات'}
						</button>
					</div>
				{/if}

			{:else if activeTab === 'webhooks'}
				<!-- Webhooks Tab -->
				<div class="space-y-6">
					<div class="flex items-center justify-between">
						<div>
							<h2 class="text-lg font-bold" style="color: var(--color-midnight-sky);">مدیریت وب‌هوک‌ها</h2>
							<p class="text-sm mt-1" style="color: var(--color-mystic-sea);">وب‌هوک‌ها برای دریافت اعلان‌های رویدادها در سیستم خارجی استفاده می‌شوند</p>
						</div>
						<button onclick={openCreateModal} class="sky-btn sky-btn-primary flex items-center gap-2">
							<svg width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
							ایجاد وب‌هوک
						</button>
					</div>

					{#if webhooksLoading}
						<div class="flex items-center justify-center py-12">
							<svg class="sky-spinner lg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--color-crystal-clear);"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/></svg>
						</div>
					{:else if webhooks.length === 0}
						<div class="sky-empty" style="background: var(--color-eternal-snow); border-radius: 12px;">
							<div class="sky-empty-icon"><svg width="48" height="48" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24" style="color: var(--color-muted-mountain);"><path d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"/></svg></div>
							<p class="sky-empty-title">هیچ وب‌هوکی ایجاد نشده است</p>
							<button onclick={openCreateModal} class="sky-btn sky-btn-ghost">ایجاد اولین وب‌هوک</button>
						</div>
					{:else}
						<div class="space-y-4">
							{#each webhooks as webhook (webhook.id)}
								<div class="sky-card overflow-hidden">
									<div class="p-4">
										<div class="flex items-start justify-between gap-4">
											<div class="flex-1 min-w-0">
												<div class="flex items-center gap-3">
													<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium {webhook.is_active ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'}">
														{webhook.is_active ? 'فعال' : 'غیرفعال'}
													</span>
													<span class="text-xs text-gray-500">
														{formatDate(webhook.created_at)}
													</span>
												</div>
												<p class="mt-2 text-sm font-mono text-gray-900 truncate" title={webhook.url}>
													{webhook.url}
												</p>
												<div class="mt-2 flex flex-wrap gap-1.5">
													{#each webhook.events as event}
														<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-800">
															{getEventLabel(event)}
														</span>
													{/each}
												</div>
											</div>
											<div class="flex items-center gap-2">
												<button
													onclick={() => testWebhook(webhook.id)}
													class="p-2 text-gray-400 hover:text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
													title="تست وب‌هوک"
												>
													<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
													</svg>
												</button>
												<button
													onclick={() => openEditModal(webhook)}
													class="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
													title="ویرایش"
												>
													<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
													</svg>
												</button>
												<button
													onclick={() => confirmDeleteWebhook(webhook.id)}
													class="p-2 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
													title="حذف"
												>
													<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
													</svg>
												</button>
											</div>
										</div>

										{#if webhookTested === webhook.id}
											<div class="mt-3 p-3 bg-green-50 border border-green-200 rounded-lg">
												<p class="text-sm text-green-700">رویداد تست با موفقیت ارسال شد</p>
											</div>
										{/if}

										<!-- Delivery logs toggle -->
										<button
											onclick={() => loadDeliveries(webhook.id)}
											class="mt-3 text-sm text-blue-600 hover:text-blue-700 font-medium flex items-center gap-1"
										>
											<svg class="w-4 h-4 transition-transform {selectedWebhookId === webhook.id && showDeliveries ? 'rotate-90' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
											</svg>
											مشاهده لاگ ارسال
										</button>
									</div>

									<!-- Delivery logs -->
									{#if selectedWebhookId === webhook.id && showDeliveries}
										<div class="border-t bg-gray-50 p-4">
											<h4 class="text-sm font-medium text-gray-900 mb-3">لاگ ارسال</h4>
											{#if deliveriesLoading}
												<div class="flex items-center justify-center py-8">
													<div class="animate-spin h-6 w-6 border-4 border-blue-600 border-t-transparent rounded-full"></div>
												</div>
											{:else if deliveries.length === 0}
												<p class="text-sm text-gray-500 text-center py-4">هیچ ارسالی ثبت نشده است</p>
											{:else}
												<div class="space-y-2 max-h-64 overflow-y-auto">
													{#each deliveries as delivery (delivery.id)}
														<div class="bg-white p-3 rounded-lg text-sm">
															<div class="flex items-center justify-between">
																<div class="flex items-center gap-2">
																	<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium {delivery.success ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}">
																		{delivery.success ? 'موفق' : 'ناموفق'}
																	</span>
																	<span class="text-gray-600">{getEventLabel(delivery.event_type)}</span>
																</div>
																<span class="text-xs text-gray-500">{formatDate(delivery.created_at)}</span>
															</div>
															{#if delivery.status_code}
																<div class="mt-1 text-xs text-gray-500">
																	کد وضعیت: <span class="font-mono">{delivery.status_code}</span>
																	{#if delivery.retry_count > 0}
																		<span class="mr-2">| تلاش: {delivery.retry_count}</span>
																	{/if}
																</div>
															{/if}
														</div>
													{/each}
												</div>
											{/if}
										</div>
									{/if}
								</div>
							{/each}
						</div>
					{/if}
				</div>

			{:else if activeTab === 'video'}
				<!-- Video Settings -->
				{#if loading}
					<div class="flex items-center justify-center py-12">
						<svg class="sky-spinner lg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--color-crystal-clear);"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/></svg>
					</div>
				{:else}
					<div class="space-y-6">
						<div>
							<h2 class="text-lg font-bold" style="color: var(--color-midnight-sky);">تنظیمات ویدیو</h2>
							<p class="text-sm text-gray-500 mt-1">پیکربندی سرور ویدیو برای جلسات آنلاین</p>
						</div>
						<div class="divide-y">
							<!-- Allow student video -->
							<div class="py-4 flex items-center justify-between">
								<div>
									<p class="font-medium" style="color: var(--color-midnight-sky);">ارسال ویدیو توسط دانش‌آموز</p>
									<p class="text-sm mt-0.5" style="color: var(--color-mystic-sea);">اجازه ارسال ویدیو به دانش‌آموزان</p>
								</div>
								<button
									onclick={() => settings.allow_student_video = !settings.allow_student_video}
									class="relative w-11 h-6 rounded-full transition-colors {settings.allow_student_video ? 'bg-[var(--color-crystal-clear)]' : 'bg-[var(--color-muted-mountain)]'}"
								>
									<span class="absolute top-0.5 right-0.5 w-5 h-5 bg-white rounded-full transition-transform {settings.allow_student_video ? 'translate-x-[-20px]' : ''}"></span>
								</button>
							</div>
						</div>
						<div class="flex items-center justify-between pt-4 border-t">
							{#if saved}
								<span class="text-sm" style="color: var(--color-lush-meadow);">ذخیره شد</span>
							{:else}
								<span></span>
							{/if}
							<button onclick={saveSettings} disabled={saving} class="sky-btn sky-btn-primary">
								{saving ? 'در حال ذخیره...' : 'ذخیره تنظیمات'}
							</button>
						</div>
					</div>
				{/if}

			{:else if activeTab === 'security'}
				<!-- Security Settings -->
				{#if loading}
					<div class="flex items-center justify-center py-12">
						<svg class="sky-spinner lg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--color-crystal-clear);"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/></svg>
					</div>
				{:else}
					<div class="space-y-6">
						<!-- Change Admin Password -->
						<div class="sky-card p-5">
							<h2 class="text-lg font-bold mb-1" style="color: var(--color-midnight-sky);">تغییر رمز عبور</h2>
							<p class="text-sm mb-4" style="color: var(--color-mystic-sea);">رمز عبور حساب مدیر خود را تغییر دهید</p>
							{#if passwordChangeSuccess}
								<div class="mb-3 p-3 rounded-lg text-sm" style="background: rgba(64,191,127,0.1); color: var(--color-lush-meadow);">رمز عبور با موفقیت تغییر کرد</div>
							{/if}
							{#if passwordChangeError}
								<div class="mb-3 p-3 rounded-lg text-sm" style="background: rgba(224,82,82,0.1); color: var(--color-fiery-passion);">{passwordChangeError}</div>
							{/if}
							<div class="space-y-3" style="max-width: 400px;">
								<div>
									<label class="sky-label">رمز عبور فعلی</label>
									<input type="password" bind:value={currentPassword} class="sky-input" dir="ltr" />
								</div>
								<div>
									<label class="sky-label">رمز عبور جدید</label>
									<input type="password" bind:value={newPassword} class="sky-input" dir="ltr" />
								</div>
								<div>
									<label class="sky-label">تکرار رمز عبور جدید</label>
									<input type="password" bind:value={confirmPassword} class="sky-input" dir="ltr" />
								</div>
								<button onclick={changeMyPassword} disabled={passwordChangeLoading || !currentPassword || !newPassword} class="sky-btn sky-btn-primary">
									{passwordChangeLoading ? 'در حال تغییر...' : 'تغییر رمز عبور'}
								</button>
							</div>
						</div>

						<div>
							<h2 class="text-lg font-bold" style="color: var(--color-midnight-sky);">تنظیمات امنیتی</h2>
							<p class="text-sm text-gray-500 mt-1">سیاست رمز عبور و امنیت حساب کاربران</p>
						</div>
						<div class="divide-y">
							<!-- Password min length -->
							<div class="py-4 flex items-center justify-between">
								<div>
									<p class="font-medium" style="color: var(--color-midnight-sky);">حداقل طول رمز عبور</p>
									<p class="text-sm mt-0.5" style="color: var(--color-mystic-sea);">حداقل تعداد کاراکتر رمز عبور</p>
								</div>
								<input type="number" bind:value={settings.password_min_length} min="6" max="32" class="sky-input text-center" style="width: 5rem;" />
							</div>
							<!-- Require uppercase -->
							<div class="py-4 flex items-center justify-between">
								<div>
									<p class="font-medium" style="color: var(--color-midnight-sky);"> الزام حرف بزرگ</p>
									<p class="text-sm mt-0.5" style="color: var(--color-mystic-sea);">رمز عبور باید شامل حداقل یک حرف بزرگ باشد</p>
								</div>
								<button
									onclick={() => settings.password_require_uppercase = !settings.password_require_uppercase}
									class="relative w-11 h-6 rounded-full transition-colors {settings.password_require_uppercase ? 'bg-[var(--color-crystal-clear)]' : 'bg-[var(--color-muted-mountain)]'}"
								>
									<span class="absolute top-0.5 right-0.5 w-5 h-5 bg-white rounded-full transition-transform {settings.password_require_uppercase ? 'translate-x-[-20px]' : ''}"></span>
								</button>
							</div>
							<!-- Require number -->
							<div class="py-4 flex items-center justify-between">
								<div>
									<p class="font-medium" style="color: var(--color-midnight-sky);">الزام عدد</p>
									<p class="text-sm mt-0.5" style="color: var(--color-mystic-sea);">رمز عبور باید شامل حداقل یک عدد باشد</p>
								</div>
								<button
									onclick={() => settings.password_require_number = !settings.password_require_number}
									class="relative w-11 h-6 rounded-full transition-colors {settings.password_require_number ? 'bg-[var(--color-crystal-clear)]' : 'bg-[var(--color-muted-mountain)]'}"
								>
									<span class="absolute top-0.5 right-0.5 w-5 h-5 bg-white rounded-full transition-transform {settings.password_require_number ? 'translate-x-[-20px]' : ''}"></span>
								</button>
							</div>
							<!-- Require special char -->
							<div class="py-4 flex items-center justify-between">
								<div>
									<p class="font-medium" style="color: var(--color-midnight-sky);">الزام کاراکتر خاص</p>
									<p class="text-sm mt-0.5" style="color: var(--color-mystic-sea);">رمز عبور باید شامل کاراکتر خاص باشد (مانند @, #, $)</p>
								</div>
								<button
									onclick={() => settings.password_require_special = !settings.password_require_special}
									class="relative w-11 h-6 rounded-full transition-colors {settings.password_require_special ? 'bg-[var(--color-crystal-clear)]' : 'bg-[var(--color-muted-mountain)]'}"
								>
									<span class="absolute top-0.5 right-0.5 w-5 h-5 bg-white rounded-full transition-transform {settings.password_require_special ? 'translate-x-[-20px]' : ''}"></span>
								</button>
							</div>
							<!-- Session timeout -->
							<div class="py-4 flex items-center justify-between">
								<div>
									<p class="font-medium" style="color: var(--color-midnight-sky);">مدت زمان نشست (دقیقه)</p>
									<p class="text-sm mt-0.5" style="color: var(--color-mystic-sea);">مدت زمان قبل از خودکار خروج کاربر</p>
								</div>
								<input type="number" bind:value={settings.session_timeout_minutes} min="5" max="1440" class="sky-input text-center" style="width: 5rem;" />
							</div>
							<!-- Max login attempts -->
							<div class="py-4 flex items-center justify-between">
								<div>
								<p class="font-medium" style="color: var(--color-midnight-sky);">حداکثر تلاش ورود</p>
								<p class="text-sm mt-0.5" style="color: var(--color-mystic-sea);">تعداد حداکثر تلاش ناموفق قبل از قفل شدن حساب</p>
								</div>
								<input type="number" bind:value={settings.max_login_attempts} min="3" max="20" class="sky-input text-center" style="width: 5rem;" />
							</div>
							<!-- Lockout duration -->
							<div class="py-4 flex items-center justify-between">
								<div>
									<p class="font-medium" style="color: var(--color-midnight-sky);">مدت قفل حساب (دقیقه)</p>
									<p class="text-sm mt-0.5" style="color: var(--color-mystic-sea);">مدت زمان قفل شدن حساب پس از تلاش‌های ناموفق</p>
								</div>
								<input type="number" bind:value={settings.lockout_duration_minutes} min="5" max="1440" class="sky-input text-center" style="width: 5rem;" />
							</div>
							<!-- Require 2FA -->
							<div class="py-4 flex items-center justify-between">
								<div>
									<p class="font-medium" style="color: var(--color-midnight-sky);">الزام احراز هویت دو مرحله‌ای</p>
									<p class="text-sm mt-0.5" style="color: var(--color-mystic-sea);">فعال‌سازی اجباری ۲FA برای تمام کاربران</p>
								</div>
								<button
									onclick={() => settings.require_2fa = !settings.require_2fa}
									class="relative w-11 h-6 rounded-full transition-colors {settings.require_2fa ? 'bg-[var(--color-crystal-clear)]' : 'bg-[var(--color-muted-mountain)]'}"
								>
									<span class="absolute top-0.5 right-0.5 w-5 h-5 bg-white rounded-full transition-transform {settings.require_2fa ? 'translate-x-[-20px]' : ''}"></span>
								</button>
							</div>
						</div>
						<div class="flex items-center justify-between pt-4 border-t">
							{#if saved}
								<span class="text-sm" style="color: var(--color-lush-meadow);">ذخیره شد</span>
							{:else}
								<span></span>
							{/if}
							<button onclick={saveSettings} disabled={saving} class="sky-btn sky-btn-primary">
								{saving ? 'در حال ذخیره...' : 'ذخیره تنظیمات'}
							</button>
						</div>
					</div>
				{/if}

			{:else if activeTab === 'email'}
				<!-- Email Settings -->
				{#if loading}
					<div class="flex items-center justify-center py-12">
						<svg class="sky-spinner lg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--color-crystal-clear);"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/></svg>
					</div>
				{:else}
					<div class="space-y-6">
						<div>
							<h2 class="text-lg font-bold" style="color: var(--color-midnight-sky);">تنظیمات ایمیل</h2>
							<p class="text-sm text-gray-500 mt-1">پیکربندی سرور SMTP برای ارسال ایمیل</p>
						</div>
						<div class="divide-y">
							<!-- SMTP Enabled -->
							<div class="py-4 flex items-center justify-between">
								<div>
									<p class="font-medium" style="color: var(--color-midnight-sky);">فعال‌سازی ایمیل</p>
									<p class="text-sm mt-0.5" style="color: var(--color-mystic-sea);">فعال‌سازی سرویس ارسال ایمیل</p>
								</div>
								<button
									onclick={() => settings.smtp_enabled = !settings.smtp_enabled}
									class="relative w-11 h-6 rounded-full transition-colors {settings.smtp_enabled ? 'bg-[var(--color-crystal-clear)]' : 'bg-[var(--color-muted-mountain)]'}"
								>
									<span class="absolute top-0.5 right-0.5 w-5 h-5 bg-white rounded-full transition-transform {settings.smtp_enabled ? 'translate-x-[-20px]' : ''}"></span>
								</button>
							</div>
							<!-- SMTP Host -->
							<div class="py-4 flex items-center justify-between">
								<div>
									<p class="font-medium" style="color: var(--color-midnight-sky);">آدرس سرور SMTP</p>
									<p class="text-sm mt-0.5" style="color: var(--color-mystic-sea);">آدرس سرور پست الکترونیک</p>
								</div>
								<input type="text" bind:value={settings.smtp_host} placeholder="smtp.gmail.com" dir="ltr" class="sky-input" style="width: 16rem;" />
							</div>
							<!-- SMTP Port -->
							<div class="py-4 flex items-center justify-between">
								<div>
									<p class="font-medium" style="color: var(--color-midnight-sky);">پورت SMTP</p>
									<p class="text-sm mt-0.5" style="color: var(--color-mystic-sea);">پورت اتصال به سرور SMTP</p>
								</div>
								<input type="number" bind:value={settings.smtp_port} min="1" max="65535" class="sky-input text-center" style="width: 6rem;" />
							</div>
							<!-- SMTP Username -->
							<div class="py-4 flex items-center justify-between">
								<div>
									<p class="font-medium" style="color: var(--color-midnight-sky);">نام کاربری SMTP</p>
									<p class="text-sm mt-0.5" style="color: var(--color-mystic-sea);">نام کاربری احراز هویت SMTP</p>
								</div>
								<input type="text" bind:value={settings.smtp_username} placeholder="your@email.com" dir="ltr" class="sky-input" style="width: 16rem;" />
							</div>
							<!-- SMTP Password -->
							<div class="py-4 flex items-center justify-between">
								<div>
									<p class="font-medium" style="color: var(--color-midnight-sky);">رمز عبور SMTP</p>
									<p class="text-sm mt-0.5" style="color: var(--color-mystic-sea);">رمز عبور احراز هویت SMTP</p>
								</div>
								<input type="password" bind:value={settings.smtp_password} placeholder="••••••••" dir="ltr" class="sky-input" style="width: 16rem;" />
							</div>
							<!-- SMTP From -->
							<div class="py-4 flex items-center justify-between">
								<div>
									<p class="font-medium" style="color: var(--color-midnight-sky);">آدرس فرستنده</p>
									<p class="text-sm mt-0.5" style="color: var(--color-mystic-sea);">آدرس ایمیل نمایش داده شده به عنوان فرستنده</p>
								</div>
								<input type="email" bind:value={settings.smtp_from} placeholder="noreply@iroom.ir" dir="ltr" class="sky-input" style="width: 16rem;" />
							</div>
						</div>

						<!-- Test Email -->
						<div class="flex items-center gap-3 pt-4 border-t">
							<button onclick={testEmail} disabled={emailTesting || !settings.smtp_enabled} class="sky-btn sky-btn-secondary flex items-center gap-2">
								{#if emailTesting}
									<svg class="sky-spinner sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/></svg>
									در حال ارسال...
								{:else}
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
									</svg>
									ارسال ایمیل تست
								{/if}
							</button>
							{#if emailTestResult === 'success'}
								<span class="text-sm text-green-600">ایمیل تست با موفقیت ارسال شد</span>
							{:else if emailTestResult === 'error'}
								<span class="text-sm text-red-600">خطا در ارسال ایمیل تست</span>
							{/if}
						</div>

						<div class="flex items-center justify-between pt-4 border-t">
							{#if saved}
								<span class="text-sm" style="color: var(--color-lush-meadow);">ذخیره شد</span>
							{:else}
								<span></span>
							{/if}
							<button onclick={saveSettings} disabled={saving} class="sky-btn sky-btn-primary">
								{saving ? 'در حال ذخیره...' : 'ذخیره تنظیمات'}
							</button>
						</div>
					</div>
				{/if}

			{:else if activeTab === 'api'}
				<!-- API Settings -->
				{#if loading}
					<div class="flex items-center justify-center py-12">
						<svg class="sky-spinner lg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--color-crystal-clear);"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/></svg>
					</div>
				{:else}
					<div class="space-y-6">
						<div>
							<h2 class="text-lg font-bold" style="color: var(--color-midnight-sky);">تنظیمات API خارجی</h2>
							<p class="text-sm text-gray-500 mt-1">مدیریت کلید API و مستندات اندپوینت‌ها</p>
						</div>

						<!-- API Key -->
						<div class="bg-gray-50 rounded-xl p-4">
							<h3 class="font-medium text-gray-900 mb-3">کلید API خارجی</h3>
							<div class="flex items-center gap-3">
								<input type="text" bind:value={settings.external_api_key} placeholder="کلید API را وارد کنید" dir="ltr" class="sky-input flex-1 font-mono" />
							</div>
							<p class="text-xs text-gray-500 mt-2">این کلید برای احراز هویت درخواست‌های API خارجی استفاده می‌شود</p>
						</div>

						<!-- API Endpoints Documentation -->
						<div>
							<h3 class="font-medium text-gray-900 mb-3">اندپوینت‌های API</h3>
							<div class="bg-gray-50 rounded-xl divide-y">
								<div class="p-4">
									<div class="flex items-center gap-2">
										<span class="px-2 py-0.5 rounded text-xs font-bold bg-green-100 text-green-800">POST</span>
										<code class="text-sm font-mono text-gray-900">/api/v1/external/users</code>
									</div>
									<p class="text-sm text-gray-500 mt-1 mr-16">ایجاد کاربر جدید از طریق API خارجی</p>
								</div>
								<div class="p-4">
									<div class="flex items-center gap-2">
										<span class="px-2 py-0.5 rounded text-xs font-bold bg-green-100 text-green-800">POST</span>
										<code class="text-sm font-mono text-gray-900">/api/v1/external/classes</code>
									</div>
									<p class="text-sm text-gray-500 mt-1 mr-16">ایجاد کلاس جدید از طریق API خارجی</p>
								</div>
								<div class="p-4">
									<div class="flex items-center gap-2">
										<span class="px-2 py-0.5 rounded text-xs font-bold bg-green-100 text-green-800">POST</span>
										<code class="text-sm font-mono text-gray-900">/api/v1/external/sessions</code>
									</div>
									<p class="text-sm text-gray-500 mt-1 mr-16">ایجاد جلسه جدید از طریق API خارجی</p>
								</div>
								<div class="p-4">
									<div class="flex items-center gap-2">
										<span class="px-2 py-0.5 rounded text-xs font-bold bg-blue-100 text-blue-800">GET</span>
										<code class="text-sm font-mono text-gray-900">/api/v1/external/status</code>
									</div>
									<p class="text-sm text-gray-500 mt-1 mr-16">دریافت وضعیت سیستم</p>
								</div>
								<div class="p-4">
									<div class="flex items-center gap-2">
										<span class="px-2 py-0.5 rounded text-xs font-bold bg-blue-100 text-blue-800">GET</span>
										<code class="text-sm font-mono text-gray-900">/api/v1/external/stats</code>
									</div>
									<p class="text-sm text-gray-500 mt-1 mr-16">دریافت آمار سیستم</p>
								</div>
							</div>
						</div>

						<!-- Rate Limit Info -->
						<div class="bg-amber-50 border border-amber-200 rounded-xl p-4">
							<div class="flex items-start gap-3">
								<svg class="w-5 h-5 text-amber-600 mt-0.5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z" />
								</svg>
								<div>
									<p class="font-medium text-amber-800">محدودیت نرخ درخواست</p>
									<p class="text-sm text-amber-700 mt-1">حداکثر ۶۰ درخواست در دقیقه برای هر کلید API مجاز است. در صورت تجاوز از این حد، درخواست‌ها با خطای 429 مواجه می‌شوند.</p>
								</div>
							</div>
						</div>

						<div class="flex items-center justify-between pt-4 border-t">
							{#if saved}
								<span class="text-sm" style="color: var(--color-lush-meadow);">ذخیره شد</span>
							{:else}
								<span></span>
							{/if}
							<button onclick={saveSettings} disabled={saving} class="sky-btn sky-btn-primary">
								{saving ? 'در حال ذخیره...' : 'ذخیره تنظیمات'}
							</button>
						</div>
					</div>
				{/if}

			{/if}
		</div>
	</div>
</div>

<!-- Create/Edit Webhook Modal -->
{#if showWebhookModal}
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div class="modal-overlay" role="button" tabindex="-1" onclick={closeWebhookModal}>
		<div class="modal-content" onclick={(e) => e.stopPropagation()}>
			<div class="sky-modal-header">
				<h2>{editingWebhook ? 'ویرایش وب‌هوک' : 'ایجاد وب‌هوک جدید'}</h2>
				<button onclick={closeWebhookModal} class="sky-btn-icon"><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg></button>
			</div>
			<div class="sky-modal-body space-y-4">
				<!-- URL Input -->
				<div>
					<label class="sky-label">آدرس URL</label>
					<input type="url" bind:value={webhookForm.url} placeholder="https://example.com/webhook" class="sky-input" dir="ltr" />
					<p class="mt-1 text-xs" style="color: var(--color-moonlit-mist);">آدرس سروری که رویدادها به آن ارسال می‌شوند</p>
				</div>

				<!-- Events Selection -->
				<div>
					<label class="sky-label">رویدادها</label>
					<div class="space-y-2">
						{#each Object.entries(WEBHOOK_EVENTS) as [eventKey, eventLabel]}
							<label class="flex items-center gap-3 p-3 rounded-lg cursor-pointer transition-colors" style="border: 1px solid {webhookForm.events.includes(eventKey) ? 'var(--color-crystal-clear)' : 'var(--color-zen-garden)'}; background: {webhookForm.events.includes(eventKey) ? 'var(--color-polar-ice)' : 'transparent'};">
								<input type="checkbox" checked={webhookForm.events.includes(eventKey)} onchange={() => toggleEvent(eventKey as WebhookEventType)} class="w-4 h-4 rounded" style="accent-color: var(--color-crystal-clear);" />
								<div>
									<span class="text-sm font-medium" style="color: var(--color-midnight-sky);">{eventLabel}</span>
									<span class="text-xs font-mono mr-2" style="color: var(--color-moonlit-mist);">({eventKey})</span>
								</div>
							</label>
						{/each}
					</div>
				</div>

				<!-- Active Toggle -->
				<div class="flex items-center justify-between p-3 rounded-lg" style="border: 1px solid var(--color-zen-garden);">
					<div>
						<span class="text-sm font-medium" style="color: var(--color-midnight-sky);">وضعیت فعال</span>
						<p class="text-xs" style="color: var(--color-moonlit-mist);">وب‌هوک فقط در حالت فعال رویدادها را ارسال می‌کند</p>
					</div>
					<button onclick={() => webhookActive = !webhookActive} class="relative w-11 h-6 rounded-full transition-colors {webhookActive ? 'bg-[var(--color-crystal-clear)]' : 'bg-[var(--color-muted-mountain)]'}">
						<span class="absolute top-0.5 right-0.5 w-5 h-5 bg-white rounded-full transition-transform {webhookActive ? 'translate-x-[-20px]' : ''}"></span>
					</button>
				</div>
			</div>
			<div class="sky-modal-footer">
				<button onclick={closeWebhookModal} class="sky-btn sky-btn-secondary">انصراف</button>
				<button onclick={saveWebhook} disabled={webhookSaving || !webhookForm.url || webhookForm.events.length === 0} class="sky-btn sky-btn-primary">
					{webhookSaving ? 'در حال ذخیره...' : (editingWebhook ? 'بروزرسانی' : 'ایجاد')}
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- Delete Confirmation Modal -->
<ConfirmModal
	show={showDeleteConfirm}
	title="حذف وب‌هوک"
	message="آیا از حذف این وب‌هوک اطمینان دارید؟ این عمل قابل بازگشت نیست."
	onConfirm={deleteWebhook}
	onCancel={() => { showDeleteConfirm = false; webhookToDelete = null; }}
/>

<style>
	@keyframes slide-up {
		from { transform: translateY(20px); opacity: 0; }
		to { transform: translateY(0); opacity: 1; }
	}
	.animate-slide-up {
		animation: slide-up 0.2s ease-out;
	}
</style>
