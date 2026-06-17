export type UserRole = 'owner' | 'admin' | 'presenter' | 'teacher' | 'student' | 'user';

export interface Participant {
	id: string;
	name: string;
	role: UserRole;
	isSpeaking: boolean;
	hasVideo: boolean;
	hasAudio: boolean;
	hasScreen: boolean;
	hasWhiteboard: boolean;
	handRaised: boolean;
	isLocal?: boolean;
	isDisconnected?: boolean;
	janusId?: number;
}

export interface ChatMessage {
	id: string;
	sender: string;
	senderId?: string;
	content: string;
	time: string;
	isOwn: boolean;
	replyTo?: {
		sender: string;
		content: string;
	};
	isEdited?: boolean;
	isPinned?: boolean;
}

export interface ClassroomState {
	sessionId: string | null;
	isOpen: boolean;
	isConnected: boolean;
	connectionState: 'disconnected' | 'connecting' | 'connected';
	audioEnabled: boolean;
	videoEnabled: boolean;
	screenSharing: boolean;
	isRecording: boolean;
	activeView: 'video' | 'whiteboard' | 'screenshare' | 'files';
	participants: Participant[];
	chatMessages: ChatMessage[];
	unreadChatCount: number;
	elapsedSeconds: number;
	showUsersPanel: boolean;
	showChatPanel: boolean;
	showAppMenu: boolean;
}

export const ROLE_HIERARCHY: Record<UserRole, number> = {
	owner: 0,
	admin: 1,
	teacher: 2,
	presenter: 3,
	user: 4,
	student: 4,
};

export const ROLE_LABELS: Record<UserRole, string> = {
	owner: 'مالک',
	admin: 'مدیر',
	teacher: 'مدرس',
	presenter: 'ارائه‌دهنده',
	user: 'کاربر',
	student: 'دانش‌آموز',
};

export function canMuteUser(currentRole: UserRole, targetRole: UserRole): boolean {
	return ROLE_HIERARCHY[currentRole] < ROLE_HIERARCHY[targetRole];
}

export function canKickUser(currentRole: UserRole, targetRole: UserRole): boolean {
	return currentRole === 'owner' || (currentRole === 'admin' && ROLE_HIERARCHY[targetRole] > ROLE_HIERARCHY.admin);
}
