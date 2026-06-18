/**
 * Classroom Types — Types for the real-time classroom experience.
 *
 * Includes:
 *   - UserRole: Permission levels in the classroom (owner > admin > operator > presenter > student)
 *   - Participant: Connected user with media state
 *   - ChatMessage: Real-time chat message
 *   - ClassroomState: Full classroom UI state
 *   - ROLE_PERMISSIONS: What each role can do (mic, webcam, screen share, etc.)
 *   - ROLE_HIERARCHY: Role ranking for permission checks
 *   - ROLE_LABELS: Persian labels for roles
 *   - canMuteUser/canKickUser: Permission check functions
 */

export type UserRole = 'owner' | 'admin' | 'operator' | 'presenter' | 'teacher' | 'student' | 'user';

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
	operator: 2,
	teacher: 3,
	presenter: 4,
	user: 5,
	student: 5,
};

export const ROLE_LABELS: Record<UserRole, string> = {
	owner: 'مالک',
	admin: 'مدیر',
	operator: 'اپراتور',
	teacher: 'مدرس',
	presenter: 'ارائه‌دهنده',
	user: 'کاربر',
	student: 'دانش‌آموز',
};

export const ROLE_PERMISSIONS: Record<UserRole, { canMic: boolean; canWebcam: boolean; canScreenShare: boolean; canWhiteboard: boolean; canHandRaise: boolean; canChat: boolean; canWatch: boolean }> = {
	owner:     { canMic: true, canWebcam: true, canScreenShare: true, canWhiteboard: true, canHandRaise: true, canChat: true, canWatch: true },
	admin:     { canMic: true, canWebcam: true, canScreenShare: true, canWhiteboard: true, canHandRaise: true, canChat: true, canWatch: true },
	operator:  { canMic: true, canWebcam: true, canScreenShare: true, canWhiteboard: true, canHandRaise: true, canChat: true, canWatch: true },
	teacher:   { canMic: true, canWebcam: true, canScreenShare: true, canWhiteboard: true, canHandRaise: true, canChat: true, canWatch: true },
	presenter: { canMic: true, canWebcam: true, canScreenShare: true, canWhiteboard: false, canHandRaise: true, canChat: true, canWatch: true },
	user:      { canMic: false, canWebcam: false, canScreenShare: false, canWhiteboard: false, canHandRaise: true, canChat: true, canWatch: true },
	student:   { canMic: false, canWebcam: false, canScreenShare: false, canWhiteboard: false, canHandRaise: true, canChat: true, canWatch: true },
};

export function canMuteUser(currentRole: UserRole, targetRole: UserRole): boolean {
	return ROLE_HIERARCHY[currentRole] < ROLE_HIERARCHY[targetRole];
}

export function canKickUser(currentRole: UserRole, targetRole: UserRole): boolean {
	return currentRole === 'owner' || (currentRole === 'admin' && ROLE_HIERARCHY[targetRole] > ROLE_HIERARCHY.admin);
}
