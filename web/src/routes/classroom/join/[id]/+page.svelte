<!--
  Classroom Join Page — Guest login for joining a class session.
  
  This is the entry point for students/parents joining a class.
  Teacher sends this URL to students. Students enter their name and join.
  
  Flow:
    1. Teacher shares /classroom/join/:sessionId with students
    2. Student enters display name
    3. Calls POST /auth/guest-login to get JWT token
    4. Redirects to /classroom/popup/:sessionId
  
  Route: /classroom/join/:sessionId
  Auth: Creates guest account (no existing account needed)
-->
<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';
	import { auth } from '$lib/stores';
	import { onMount } from 'svelte';
	import type { User, Tokens } from '$lib/types';

	let displayName = $state('');
	let loading = $state(false);
	let error = $state('');
	let session = $state<any>(null);

	const sessionId = $derived(page.params.id!);

	onMount(async () => {
		const res = await api.get(`/sessions/${sessionId}/info`);
		if (res.success) session = res.data!;
	});

	async function handleJoin() {
		if (!displayName.trim()) {
			error = 'لطفاً نام خود را وارد کنید';
			return;
		}
		loading = true;
		error = '';

		const res = await api.post<{ user: User; tokens: Tokens }>('/auth/guest-login', {
			session_id: Number(sessionId),
			display_name: displayName.trim()
		});

		if (!res.success) {
			error = res.error || 'خطا در ورود مهمان';
			loading = false;
			return;
		}

		auth.login(res.data!.user, res.data!.tokens);
		goto(`/classroom/popup/${sessionId}`);
	}
</script>

<div class="min-h-screen flex items-center justify-center px-4" style="background: linear-gradient(135deg, #0b1120 0%, #1a1a2e 50%, #0d1b2a 100%);">
	<div class="w-full max-w-[400px] rounded-2xl p-6" style="background: #16213e; border: 1px solid #2a2a4a; box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);">
		<!-- Logo -->
		<div class="text-center mb-6">
			<div class="w-14 h-14 rounded-xl mx-auto mb-3 flex items-center justify-center text-white font-bold text-xl" style="background: linear-gradient(135deg, #23b9d7, #004ff2);">
				آ
			</div>
			<h1 class="text-xl font-bold" style="color: #eaeaea;">آی‌روم</h1>
		</div>

		{#if session}
			<!-- Session info -->
			<div class="mb-4 p-3 rounded-lg text-center" style="background: rgba(35, 185, 215, 0.1); border: 1px solid rgba(35, 185, 215, 0.2);">
				<p class="text-sm font-medium" style="color: #23b9d7;">{session.title}</p>
				<p class="text-xs mt-1" style="color: #94a3b8;">
					{#if session.status === 'live'}
						<span style="color: #40bf7f;">●</span> در حال برگزاری
					{:else if session.status === 'scheduled'}
						<span style="color: #f59e0b;">●</span> برنامه‌ریزی شده
					{:else}
						<span style="color: #e05252;">●</span> پایان یافته
					{/if}
				</p>
			</div>
		{/if}

		<!-- Error -->
		{#if error}
			<div class="mb-4 px-4 py-3 rounded-lg text-sm text-center" style="background: rgba(224, 82, 82, 0.08); color: #e05252; border: 1px solid rgba(224, 82, 82, 0.2);">
				{error}
			</div>
		{/if}

		<!-- Name input -->
		<form onsubmit={(e) => { e.preventDefault(); handleJoin(); }} class="space-y-3">
			<div>
				<label class="block text-xs font-medium mb-1.5" style="color: #94a3b8;">نام شما</label>
				<input type="text" bind:value={displayName} placeholder="نام خود را وارد کنید" dir="auto" required
					class="w-full px-4 py-2.5 rounded-lg text-sm outline-none transition-colors"
					style="border: 1px solid #2a2a4a; color: #eaeaea; background: #0f3460; font-family: 'Vazirmatn', system-ui, sans-serif;"
					onfocus={(e) => e.currentTarget.style.borderColor = '#23b9d7'}
					onblur={(e) => e.currentTarget.style.borderColor = '#2a2a4a'}
				/>
			</div>

			<button type="submit" disabled={loading || (session && session.status !== 'live')}
				class="w-full py-2.5 rounded-lg text-sm font-semibold text-white transition-colors"
				style="background: {(session && session.status === 'live') ? '#23b9d7' : '#4a5568'}; font-family: 'Vazirmatn', system-ui, sans-serif; {loading ? 'opacity: 0.6; cursor: not-allowed;' : ''}"
			>
				{loading ? 'در حال پیوستن...' : (session && session.status === 'live' ? 'پیوستن به کلاس' : 'جلسه در حال برگزاری نیست')}
			</button>
		</form>

		<!-- Footer -->
		<div class="mt-4 text-center">
			<a href="/" class="text-xs" style="color: #6790a0;">ورود به پنل مدیریت</a>
		</div>
	</div>
</div>
