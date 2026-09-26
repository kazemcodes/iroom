<script lang="ts">
	import { api } from '$lib/api';
	import { onMount } from 'svelte';
	import ConfirmModal from '$lib/components/ConfirmModal.svelte';
	import type { Webhook, WebhookDelivery, CreateWebhookRequest } from '$lib/types';
	import { WEBHOOK_EVENTS, type WebhookEventType } from '$lib/types';

	let activeTab = $state('general');
	const tabs = [
		{ id: 'general', label: 'عمومی' },
		{ id: 'api', label: 'API' },
		{ id: 'webhooks', label: 'وب‌هوک‌ها' },
	];

	let settings = $state({
		max_users_per_room: 100,
		recording_enabled: true,
		maintenance_mode: false,
		max_file_size_mb: 50,
		session_auto_end_minutes: 120,
		external_api_key: '',
	});
	let loading = $state(true);
	let saving = $state(false);
	let saved = $state(false);

	let webhooks = $state<Webhook[]>([]);
	let webhooksLoading = $state(true);
	let showWebhookModal = $state(false);
	let editingWebhook = $state<Webhook | null>(null);
	let webhookSaving = $state(false);
	let webhookForm = $state<CreateWebhookRequest>({ url: '', events: [] });
	let webhookActive = $state(true);
	let deliveries = $state<WebhookDelivery[]>([]);
	let deliveriesLoading = $state(false);
	let showDeliveries = $state(false);
	let showDeleteConfirm = $state(false);
	let webhookToDelete = $state<number | null>(null);

	// Password change
	let currentPassword = $state('');
	let newPassword = $state('');
	let confirmPassword = $state('');
	let passwordChangeLoading = $state(false);
	let passwordChangeSuccess = $state(false);
	let passwordChangeError = $state('');
	let apiCopied = $state(false);

	onMount(async () => {
		const res = await api.get<any>('/admin/settings');
		if (res.success && res.data) settings = { ...settings, ...res.data };
		loading = false;
		await loadWebhooks();
	});

	async function saveSettings() {
		saving = true; saved = false;
		const res = await api.put('/admin/settings', settings);
		if (res.success) { saved = true; setTimeout(() => saved = false, 3000); }
		saving = false;
	}

	async function changeMyPassword() {
		if (newPassword !== confirmPassword) { passwordChangeError = 'رمز عبور جدید و تکرار آن مطابقت ندارند'; return; }
		if (newPassword.length < 6) { passwordChangeError = 'رمز عبور باید حداقل ۶ کاراکتر باشد'; return; }
		passwordChangeLoading = true; passwordChangeError = ''; passwordChangeSuccess = false;
		const res = await api.post('/auth/change-password', { old_password: currentPassword, new_password: newPassword });
		if (res.success) {
			passwordChangeSuccess = true; currentPassword = ''; newPassword = ''; confirmPassword = '';
			setTimeout(() => passwordChangeSuccess = false, 3000);
		} else { passwordChangeError = res.error || 'خطا در تغییر رمز عبور'; }
		passwordChangeLoading = false;
	}

	// Webhooks
	async function loadWebhooks() {
		webhooksLoading = true;
		const res = await api.get<Webhook[]>('/admin/webhooks');
		if (res.success && res.data) webhooks = res.data;
		webhooksLoading = false;
	}

	async function saveWebhook() {
		webhookSaving = true;
		if (editingWebhook) {
			await api.put(`/admin/webhooks/${editingWebhook.id}`, { url: webhookForm.url, events: webhookForm.events, is_active: webhookActive });
		} else {
			await api.post('/admin/webhooks', { url: webhookForm.url, events: webhookForm.events });
		}
		webhookSaving = false; showWebhookModal = false; editingWebhook = null;
		await loadWebhooks();
	}

	function editWebhook(wh: Webhook) {
		editingWebhook = wh;
		webhookForm = { url: wh.url, events: [...wh.events] };
		webhookActive = wh.is_active;
		showWebhookModal = true;
	}

	function confirmDeleteWebhook(id: number) { webhookToDelete = id; showDeleteConfirm = true; }

	async function deleteWebhook() {
		if (!webhookToDelete) return;
		await api.delete(`/admin/webhooks/${webhookToDelete}`);
		showDeleteConfirm = false; webhookToDelete = null;
		await loadWebhooks();
	}

	async function testWebhook(id: number) {
		await api.post(`/admin/webhooks/${id}/test`);
		await loadWebhooks();
	}

	function toggleEvent(event: WebhookEventType) {
		const idx = webhookForm.events.indexOf(event);
		if (idx >= 0) webhookForm.events.splice(idx, 1); else webhookForm.events.push(event);
		webhookForm = { ...webhookForm };
	}

	function toPersian(n: number) { return n.toLocaleString('fa-IR'); }

	function generateApiKey(): string {
		const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
		let key = 'irk_';
		for (let i = 0; i < 32; i++) {
			key += chars.charAt(Math.floor(Math.random() * chars.length));
		}
		return key;
	}
</script>

<div style="max-width:800px;margin:0 auto;" class="space-y-5">
	<div>
		<h1 class="sky-page-title">تنظیمات سیستم</h1>
		<p class="sky-page-subtitle">تنظیمات کلی پلتفرم کلاس آنلاین</p>
	</div>

	<div class="sky-card overflow-hidden">
		<div class="sky-tabs" style="padding: 0 1.25rem; overflow-x: auto;">
			{#each tabs as tab}
				<button class="sky-tab {activeTab === tab.id ? 'active' : ''}" onclick={() => activeTab = tab.id}>{tab.label}</button>
			{/each}
		</div>

		<div class="p-6">
			{#if activeTab === 'general'}
				{#if loading}
					<div class="flex items-center justify-center py-12"><svg class="sky-spinner lg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--color-crystal-clear);"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/></svg></div>
				{:else}
					<div class="space-y-6">
						<!-- Change Admin Password -->
						<div class="sky-card p-5">
							<h2 class="text-lg font-bold mb-1" style="color: var(--color-midnight-sky);">تغییر رمز عبور</h2>
							<p class="text-sm mb-4" style="color: var(--color-mystic-sea);">رمز عبور حساب مدیر خود را تغییر دهید</p>
							{#if passwordChangeSuccess}<div class="mb-3 p-3 rounded-lg text-sm" style="background: rgba(64,191,127,0.1); color: var(--color-lush-meadow);">رمز عبور با موفقیت تغییر کرد</div>{/if}
							{#if passwordChangeError}<div class="mb-3 p-3 rounded-lg text-sm" style="background: rgba(224,82,82,0.1); color: var(--color-fiery-passion);">{passwordChangeError}</div>{/if}
							<div class="space-y-3" style="max-width: 400px;">
								<div><label class="sky-label">رمز عبور فعلی</label><input type="password" bind:value={currentPassword} class="sky-input" dir="ltr" /></div>
								<div><label class="sky-label">رمز عبور جدید</label><input type="password" bind:value={newPassword} class="sky-input" dir="ltr" /></div>
								<div><label class="sky-label">تکرار رمز عبور جدید</label><input type="password" bind:value={confirmPassword} class="sky-input" dir="ltr" /></div>
								<button onclick={changeMyPassword} disabled={passwordChangeLoading || !currentPassword || !newPassword} class="sky-btn sky-btn-primary">{passwordChangeLoading ? 'در حال تغییر...' : 'تغییر رمز عبور'}</button>
							</div>
						</div>

						<div class="divide-y">
							<div class="py-4 flex items-center justify-between">
								<div><p class="font-medium" style="color: var(--color-midnight-sky);">حداکثر کاربر در اتاق</p><p class="text-sm mt-0.5" style="color: var(--color-mystic-sea);">تعداد حداکثر شرکت‌کنندگان در هر جلسه</p></div>
								<input type="number" bind:value={settings.max_users_per_room} min="2" max="500" class="sky-input text-center" style="width: 5rem;" />
							</div>
							<div class="py-4 flex items-center justify-between">
								<div><p class="font-medium" style="color: var(--color-midnight-sky);">ضبط جلسات</p><p class="text-sm mt-0.5" style="color: var(--color-mystic-sea);">امکان ضبط جلسات توسط مدرس</p></div>
								<button onclick={() => settings.recording_enabled = !settings.recording_enabled} class="relative w-11 h-6 rounded-full transition-colors {settings.recording_enabled ? 'bg-[var(--color-crystal-clear)]' : 'bg-[var(--color-muted-mountain)]'}"><span class="absolute top-0.5 right-0.5 w-5 h-5 bg-white rounded-full transition-transform {settings.recording_enabled ? 'translate-x-[-20px]' : ''}"></span></button>
							</div>
							<div class="py-4 flex items-center justify-between">
								<div><p class="font-medium" style="color: var(--color-midnight-sky);">حداکثر حجم فایل (MB)</p><p class="text-sm mt-0.5" style="color: var(--color-mystic-sea);">حداکثر اندازه آپلود فایل</p></div>
								<input type="number" bind:value={settings.max_file_size_mb} min="1" max="500" class="sky-input text-center" style="width: 5rem;" />
							</div>
							<div class="py-4 flex items-center justify-between">
								<div><p class="font-medium" style="color: var(--color-midnight-sky);">پایان خودکار جلسه (دقیقه)</p><p class="text-sm mt-0.5" style="color: var(--color-mystic-sea);">زمان پایان خودکار پس از شروع</p></div>
								<input type="number" bind:value={settings.session_auto_end_minutes} min="30" max="480" class="sky-input text-center" style="width: 5rem;" />
							</div>
							<div class="py-4 flex items-center justify-between">
								<div><p class="font-medium" style="color: var(--color-midnight-sky);">حالت تعمیر و نگهداری</p><p class="text-sm mt-0.5" style="color: var(--color-mystic-sea);">غیرفعال‌سازی موقت سیستم</p></div>
								<button onclick={() => settings.maintenance_mode = !settings.maintenance_mode} class="relative w-11 h-6 rounded-full transition-colors {settings.maintenance_mode ? 'bg-[var(--color-danger)]' : 'bg-[var(--color-muted-mountain)]'}"><span class="absolute top-0.5 right-0.5 w-5 h-5 bg-white rounded-full transition-transform {settings.maintenance_mode ? 'translate-x-[-20px]' : ''}"></span></button>
							</div>
						</div>
						<div class="flex items-center justify-between pt-4 border-t">
							{#if saved}<span class="text-sm" style="color: var(--color-lush-meadow);">ذخیره شد</span>{:else}<span></span>{/if}
							<button onclick={saveSettings} disabled={saving} class="sky-btn sky-btn-primary">{saving ? 'در حال ذخیره...' : 'ذخیره تنظیمات'}</button>
						</div>
					</div>
				{/if}

			{:else if activeTab === 'api'}
				<div class="space-y-6">
					<div><h2 class="text-lg font-bold" style="color: var(--color-midnight-sky);">تنظیمات API خارجی</h2><p class="text-sm text-gray-500 mt-1">کلید API برای اتصال سرویس‌های خارجی</p></div>
					<div>
						<label class="sky-label">کلید API</label>
						<div class="flex items-center gap-2">
							<input type="text" bind:value={settings.external_api_key} class="sky-input flex-1" placeholder="کلید API را وارد کنید" dir="ltr" readonly />
							<button onclick={() => { navigator.clipboard.writeText(settings.external_api_key); apiCopied = true; setTimeout(() => apiCopied = false, 2000); }} class="sky-btn sky-btn-secondary whitespace-nowrap">{apiCopied ? 'کپی شد ✓' : 'کپی'}</button>
							<button onclick={() => { settings.external_api_key = generateApiKey(); }} class="sky-btn sky-btn-secondary whitespace-nowrap">تولید کلید جدید</button>
						</div>
					</div>
					<div class="flex items-center justify-between pt-4 border-t">
						{#if saved}<span class="text-sm" style="color: var(--color-lush-meadow);">ذخیره شد</span>{:else}<span></span>{/if}
						<button onclick={saveSettings} disabled={saving} class="sky-btn sky-btn-primary">{saving ? 'در حال ذخیره...' : 'ذخیره تنظیمات'}</button>
					</div>

					<!-- API Documentation -->
					<div class="pt-6 border-t">
						<h3 class="text-lg font-bold mb-4" style="color: var(--color-midnight-sky);">مستندات API</h3>
						<div class="space-y-4">
							<div class="sky-card p-4">
								<h4 class="font-bold mb-2" style="color: var(--color-midnight-sky);">احراز هویت</h4>
								<p class="text-sm mb-2" style="color: var(--color-mystic-sea);">تمام درخواست‌ها باید شامل هدر زیر باشند:</p>
								<pre class="p-3 rounded-lg text-xs overflow-x-auto" style="background: var(--color-secret-glow); color: var(--color-midnight-sky);"><code>X-API-Key: {settings.external_api_key || 'your-api-key-here'}</code></pre>
							</div>

							<div class="sky-card p-4">
								<h4 class="font-bold mb-2" style="color: var(--color-midnight-sky);">ایجاد کلاس</h4>
								<pre class="p-3 rounded-lg text-xs overflow-x-auto mb-2" style="background: var(--color-secret-glow); color: var(--color-midnight-sky);"><code>POST /api/v1/external/classes
Content-Type: application/json
X-API-Key: your-api-key

{`{
  "name": "نام کلاس",
  "description": "توضیحات",
  "teacher_id": 1
}`}</code></pre>
							</div>

							<div class="sky-card p-4">
								<h4 class="font-bold mb-2" style="color: var(--color-midnight-sky);">ایجاد جلسه</h4>
								<pre class="p-3 rounded-lg text-xs overflow-x-auto mb-2" style="background: var(--color-secret-glow); color: var(--color-midnight-sky);"><code>POST /api/v1/external/sessions
Content-Type: application/json
X-API-Key: your-api-key

{`{
  "class_id": 1,
  "title": "عنوان جلسه",
  "scheduled_at": "2024-01-01T10:00:00Z",
  "duration": 60
}`}</code></pre>
							</div>

							<div class="sky-card p-4">
								<h4 class="font-bold mb-2" style="color: var(--color-midnight-sky);">ایجاد لینک ورود</h4>
								<pre class="p-3 rounded-lg text-xs overflow-x-auto mb-2" style="background: var(--color-secret-glow); color: var(--color-midnight-sky);"><code>POST /api/v1/auth/create-login-url
Content-Type: application/json
X-API-Key: your-api-key

{`{
  "room_id": 1,
  "user_id": "user-123",
  "nickname": "نام کاربر",
  "access": 1,
  "concurrent": 1,
  "ttl": 3600
}`}</code></pre>
							</div>

							<div class="sky-card p-4">
								<h4 class="font-bold mb-2" style="color: var(--color-midnight-sky);">دریافت اطلاعات اتاق</h4>
								<pre class="p-3 rounded-lg text-xs overflow-x-auto mb-2" style="background: var(--color-secret-glow); color: var(--color-midnight-sky);"><code>GET /api/v1/rooms/slug/:slug
X-API-Key: your-api-key</code></pre>
								<p class="text-xs" style="color: var(--color-mystic-sea);">این endpoint نیازی به احراز هویت ندارد.</p>
							</div>

							<div class="sky-card p-4">
								<h4 class="font-bold mb-2" style="color: var(--color-midnight-sky);">لیست اتاق‌ها</h4>
								<pre class="p-3 rounded-lg text-xs overflow-x-auto mb-2" style="background: var(--color-secret-glow); color: var(--color-midnight-sky);"><code>GET /api/v1/admin/rooms?page=1&per_page=20
X-API-Key: your-api-key</code></pre>
							</div>

							<div class="sky-card p-4">
								<h4 class="font-bold mb-2" style="color: var(--color-midnight-sky);">لیست کاربران</h4>
								<pre class="p-3 rounded-lg text-xs overflow-x-auto mb-2" style="background: var(--color-secret-glow); color: var(--color-midnight-sky);"><code>GET /api/v1/admin/users?page=1&per_page=20
X-API-Key: your-api-key</code></pre>
							</div>
						</div>
					</div>
				</div>

			{:else if activeTab === 'webhooks'}
				<div class="space-y-4">
					<div class="flex items-center justify-between">
						<div><h2 class="text-lg font-bold" style="color: var(--color-midnight-sky);">وب‌هوک‌ها</h2><p class="text-sm text-gray-500 mt-1">اعلان خودکار رویدادها به سرویس‌های خارجی</p></div>
						<button onclick={() => { editingWebhook = null; webhookForm = { url: '', events: [] }; webhookActive = true; showWebhookModal = true; }} class="sky-btn sky-btn-primary">وب‌هوک جدید</button>
					</div>
					{#if webhooksLoading}
						<div class="flex justify-center py-8"><svg class="sky-spinner lg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--color-crystal-clear);"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/></svg></div>
					{:else if webhooks.length === 0}
						<div class="sky-card"><div class="sky-empty"><p class="sky-empty-desc">وب‌هوکی تعریف نشده</p></div></div>
					{:else}
						<div class="space-y-3">
							{#each webhooks as wh}
								<div class="sky-card p-4">
									<div class="flex items-center justify-between mb-2">
										<div class="flex items-center gap-2">
											<span class="sky-badge {wh.is_active ? 'sky-badge-success' : 'sky-badge-default'}">{wh.is_active ? 'فعال' : 'غیرفعال'}</span>
											<span class="font-mono text-sm" style="color: var(--color-midnight-sky);" dir="ltr">{wh.url}</span>
										</div>
										<div class="flex items-center gap-1">
											<button onclick={() => testWebhook(wh.id)} class="sky-btn sky-btn-ghost" style="font-size:12px;">تست</button>
											<button onclick={() => editWebhook(wh)} class="sky-btn sky-btn-ghost" style="font-size:12px;">ویرایش</button>
											<button onclick={() => confirmDeleteWebhook(wh.id)} class="sky-btn sky-btn-ghost" style="font-size:12px;color:var(--color-fiery-passion);">حذف</button>
										</div>
									</div>
									<div class="flex flex-wrap gap-1">
										{#each wh.events as event}
											<span class="sky-badge sky-badge-info" style="font-size:11px;">{WEBHOOK_EVENTS[event] || event}</span>
										{/each}
									</div>
								</div>
							{/each}
						</div>
					{/if}
				</div>
			{/if}
		</div>
	</div>
</div>

<!-- Webhook Modal -->
{#if showWebhookModal}
	<div class="modal-overlay" onclick={() => showWebhookModal = false} role="button" tabindex="-1">
		<div class="modal-content" onclick={(e) => e.stopPropagation()}>
			<div class="sky-modal-header">
				<h2>{editingWebhook ? 'ویرایش وب‌هوک' : 'وب‌هوک جدید'}</h2>
				<button onclick={() => showWebhookModal = false} class="sky-btn-icon"><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg></button>
			</div>
			<div class="sky-modal-body space-y-4">
				<div><label class="sky-label">آدرس URL</label><input type="url" bind:value={webhookForm.url} class="sky-input" placeholder="https://example.com/webhook" dir="ltr" /></div>
				<div>
					<label class="sky-label">رویدادها</label>
					<div class="flex flex-wrap gap-2 mt-1">
						{#each Object.entries(WEBHOOK_EVENTS) as [key, label]}
							<button onclick={() => toggleEvent(key as WebhookEventType)} class="sky-badge cursor-pointer transition-all {webhookForm.events.includes(key as WebhookEventType) ? 'sky-badge-success' : 'sky-badge-default'}" style="font-size:12px;">{label}</button>
						{/each}
					</div>
				</div>
				<div class="flex items-center justify-between">
					<span class="text-sm font-medium" style="color: var(--color-ocean-wave);">فعال</span>
					<button onclick={() => webhookActive = !webhookActive} class="relative w-11 h-6 rounded-full transition-colors {webhookActive ? 'bg-[var(--color-crystal-clear)]' : 'bg-[var(--color-muted-mountain)]'}"><span class="absolute top-0.5 right-0.5 w-5 h-5 bg-white rounded-full transition-transform {webhookActive ? 'translate-x-[-20px]' : ''}"></span></button>
				</div>
			</div>
			<div class="sky-modal-footer">
				<button onclick={() => showWebhookModal = false} class="sky-btn sky-btn-secondary">انصراف</button>
				<button onclick={saveWebhook} disabled={webhookSaving || !webhookForm.url} class="sky-btn sky-btn-primary">{webhookSaving ? 'در حال ذخیره...' : 'ذخیره'}</button>
			</div>
		</div>
	</div>
{/if}

<ConfirmModal bind:show={showDeleteConfirm} title="حذف وب‌هوک" message="آیا از حذف این وب‌هوک اطمینان دارید؟" onConfirm={() => { showDeleteConfirm = false; deleteWebhook(); }} onCancel={() => {}} />
