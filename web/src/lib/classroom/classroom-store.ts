import { writable, derived } from 'svelte/store';

export interface Participant {
  identity: string;
  name: string;
  role: string;
  isSpeaking: boolean;
  hasVideo: boolean;
  hasAudio: boolean;
  handRaised: boolean;
  isLocal?: boolean;
}

export interface ChatMessage {
  sender: string;
  content: string;
  time: string;
  isOwn: boolean;
}

export const classroomState = writable({
  sessionId: null as string | null,
  isOpen: false,
  isConnected: false,
  connectionState: 'disconnected' as string,
  audioEnabled: true,
  videoEnabled: false,
  screenSharing: false,
  isRecording: false,
  activeView: 'video' as 'video' | 'whiteboard' | 'screenshare',
  participants: [] as Participant[],
  chatMessages: [] as ChatMessage[],
  unreadChatCount: 0,
  elapsedSeconds: 0,
});

export const teacher = derived(classroomState, $s => $s.participants.find(p => p.role === 'teacher'));
export const activeSpeaker = derived(classroomState, $s => $s.participants.find(p => p.isSpeaking));
export const raisedHands = derived(classroomState, $s => $s.participants.filter(p => p.handRaised));
