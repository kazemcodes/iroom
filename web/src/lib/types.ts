export interface User {
	id: number;
	email: string;
	display_name: string;
	role: 'admin' | 'teacher' | 'student';
	phone: string;
	is_active: boolean;
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
