// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
declare global {
	// eslint-disable-next-line no-var
	var Janus: {
		init: (config: { debug: string | boolean; callback: () => void }) => void;
		attachMediaStream: (element: HTMLMediaElement, stream: MediaStream) => void;
		new (config: {
			server: string;
			success?: () => void;
			error?: (err: any) => void;
		}): any;
	};

	namespace App {
		// interface Error {}
		// interface Locals {}
		// interface PageData {}
		// interface PageState {}
		// interface Platform {}
	}
}

export {};
