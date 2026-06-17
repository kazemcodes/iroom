export interface User {
	id: number;
	email: string;
	display_name: string;
	role: 'admin' | 'teacher' | 'student';
	phone: string;
	is_active: boolean;
	two_factor_enabled?: boolean;
	created_at: string;
	updated_at: string;
}

export interface Class {
	id: number;
	teacher_id: number;
	name: string;
	description: string;
	color: string;
	max_students: number;
	created_at: string;
	updated_at: string;
}

export interface Session {
	id: number;
	class_id: number;
	title: string;
	scheduled_at: string;
	duration: number;
	status: 'scheduled' | 'live' | 'ended';
	livekit_room: string;
	recording_url: string;
	created_at: string;
	updated_at: string;
}

export interface Message {
	id: number;
	session_id: number;
	user_id: number;
	content: string;
	type: 'text' | 'file' | 'system';
	created_at: string;
}

export interface FileItem {
	id: number;
	session_id: number;
	uploaded_by: number;
	filename: string;
	filepath: string;
	filesize: number;
	created_at: string;
}

export interface Tokens {
	access_token: string;
	refresh_token: string;
	expires_in: number;
}

export interface APIResponse<T = any> {
	success: boolean;
	data?: T;
	error?: string;
}

export interface PaginatedResponse<T> {
	items: T[];
	total: number;
	page: number;
	per_page: number;
	total_pages: number;
}

export interface DashboardStats {
	users: number;
	classes: number;
	sessions: number;
	messages: number;
}

export interface ActivityLog {
	id: number;
	user_id: number;
	action: string;
	entity_type: string;
	entity_id: number;
	details: string;
	ip_address: string;
	created_at: string;
}

export interface Recording {
	id: number;
	session_id: number;
	uploaded_by: number;
	filename: string;
	filepath: string;
	filesize: number;
	duration: number;
	status: 'processing' | 'ready' | 'failed';
	created_at: string;
}

export interface Ticket {
	id: number;
	user_id: number;
	title: string;
	category: string;
	status: 'open' | 'answered' | 'closed';
	priority: 'low' | 'normal' | 'high' | 'urgent';
	created_at: string;
	updated_at: string;
	user_display_name: string;
	messages?: TicketMessage[];
}

export interface TicketMessage {
	id: number;
	ticket_id: number;
	user_id: number;
	content: string;
	is_admin: boolean;
	created_at: string;
	user_display_name: string;
}

export interface SessionLog {
	id: number;
	session_id: number;
	user_id: number;
	joined_at: string;
	left_at: string | null;
	duration: number;
	ip_address: string;
	user_display_name: string;
}

export interface Announcement {
	id: number;
	class_id: number | null;
	author_id: number;
	title: string;
	content: string;
	is_pinned: boolean;
	is_system_wide: boolean;
	created_at: string;
	updated_at: string;
	author_name?: string;
	is_read?: boolean;
}

export interface RecurringSession {
	id: number;
	class_id: number;
	title: string;
	day_of_week: number; // 0=شنبه, 1=یکشنبه, ..., 6=جمعه
	time: string; // HH:MM format
	duration: number; // minutes
	week_count: number;
	sessions_generated: number;
	created_at: string;
	updated_at: string;
}

export interface Webhook {
	id: number;
	user_id: number;
	url: string;
	events: string[];
	is_active: boolean;
	created_at: string;
	delivery_count?: number;
}

export interface WebhookDelivery {
	id: number;
	webhook_id: number;
	event_type: string;
	payload: string;
	status_code?: number;
	response_body?: string;
	success: boolean;
	retry_count: number;
	created_at: string;
}

export interface CreateWebhookRequest {
	url: string;
	events: string[];
}

export interface UpdateWebhookRequest {
	url?: string;
	events?: string[];
	is_active?: boolean;
}

export const WEBHOOK_EVENTS = {
	'session.started': 'شروع جلسه',
	'session.ended': 'پایان جلسه',
	'user.registered': 'ثبت‌نام کاربر',
	'ticket.created': 'ایجاد تیکت',
} as const;

export type WebhookEventType = keyof typeof WEBHOOK_EVENTS;
