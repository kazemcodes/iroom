<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import * as fabric from 'fabric';

	let { room, sessionId }: { room: any; sessionId: string } = $props();

	let canvasEl: HTMLCanvasElement;
	let canvas: fabric.Canvas;
	let activeTool = $state('pen');
	let brushColor = $state('#000000');
	let brushWidth = $state(3);

	const colors = ['#000000', '#EF4444', '#3B82F6', '#10B981', '#F59E0B', '#8B5CF6'];

	onMount(() => {
		canvas = new fabric.Canvas(canvasEl, {
			backgroundColor: '#ffffff',
			width: 800,
			height: 600,
		});

		canvas.on('object:added', (e) => {
			if (e.target && !isLocalUpdate) {
				syncToRoom(JSON.stringify(e.target.toJSON()));
			}
		});

		canvas.on('object:modified', (e) => {
			if (e.target) {
				syncToRoom(JSON.stringify({ op: 'update', data: e.target.toJSON() }));
			}
		});

		room?.on?.('dataReceived', (payload: Uint8Array) => {
			try {
				const data = JSON.parse(new TextDecoder().decode(payload));
				if (data.type === 'whiteboard') {
					handleRemoteOp(data);
				}
			} catch (e) {}
		});

		setTool('pen');
	});

	onDestroy(() => {
		canvas?.dispose();
	});

	let isLocalUpdate = false;

	function setTool(tool: string) {
		activeTool = tool;
		canvas.isDrawingMode = tool === 'pen' || tool === 'eraser';

		if (tool === 'pen') {
			canvas.freeDrawingBrush = new fabric.PencilBrush(canvas);
			canvas.freeDrawingBrush.color = brushColor;
			canvas.freeDrawingBrush.width = brushWidth;
		} else if (tool === 'eraser') {
			canvas.freeDrawingBrush = new fabric.PencilBrush(canvas);
			canvas.freeDrawingBrush.color = '#ffffff';
			canvas.freeDrawingBrush.width = 20;
		} else {
			canvas.selection = true;
		}
	}

	function setColor(color: string) {
		brushColor = color;
		if (canvas.freeDrawingBrush) {
			canvas.freeDrawingBrush.color = activeTool === 'eraser' ? '#ffffff' : color;
		}
	}

	function setWidth(w: number) {
		brushWidth = w;
		if (canvas.freeDrawingBrush) {
			canvas.freeDrawingBrush.width = w;
		}
	}

	function addShape(type: string) {
		let obj: fabric.FabricObject;
		const left = 100 + Math.random() * 200;
		const top = 100 + Math.random() * 200;

		switch (type) {
			case 'rect':
				obj = new fabric.Rect({ left, top, width: 150, height: 100, fill: 'transparent', stroke: brushColor, strokeWidth: 2 });
				break;
			case 'circle':
				obj = new fabric.Circle({ left, top, radius: 50, fill: 'transparent', stroke: brushColor, strokeWidth: 2 });
				break;
			case 'line':
				obj = new fabric.Line([left, top, left + 150, top], { stroke: brushColor, strokeWidth: 2 });
				break;
			case 'text':
				obj = new fabric.IText('متن', { left, top, fontSize: 24, fill: brushColor, fontFamily: 'Vazirmatn' });
				break;
			default:
				return;
		}

		isLocalUpdate = true;
		canvas.add(obj);
		isLocalUpdate = false;
		syncToRoom(JSON.stringify({ op: 'add', data: obj.toJSON() }));
	}

	function syncToRoom(data: string) {
		try {
			room?.localParticipant?.sendData(
				new TextEncoder().encode(JSON.stringify({ type: 'whiteboard', data: JSON.parse(data) })),
				{ reliable: true }
			);
		} catch (e) {
			// sendData not available in this livekit-client version
		}
	}

	function handleRemoteOp(msg: any) {
		if (msg.data?.op === 'add') {
			isLocalUpdate = true;
			fabric.util.enlivenObjects([msg.data.data], (objects: fabric.FabricObject[]) => {
				objects.forEach(obj => canvas.add(obj));
			});
			isLocalUpdate = false;
		} else if (msg.data?.op === 'clear') {
			canvas.clear();
			canvas.backgroundColor = '#ffffff';
			canvas.renderAll();
		} else if (msg.data?.op === 'update' && msg.data.data) {
			isLocalUpdate = true;
			fabric.util.enlivenObjects([msg.data.data], (objects: fabric.FabricObject[]) => {
				objects.forEach(obj => {
					const existing = canvas.getObjects().find((o: any) => o.id === obj.id);
					if (existing) {
						existing.set(obj);
						canvas.renderAll();
					}
				});
			});
			isLocalUpdate = false;
		}
	}

	function clearCanvas() {
		canvas.clear();
		canvas.backgroundColor = '#ffffff';
		canvas.renderAll();
		syncToRoom(JSON.stringify({ op: 'clear' }));
	}

	function undo() {
		const objects = canvas.getObjects();
		if (objects.length > 0) {
			const last = objects[objects.length - 1];
			canvas.remove(last);
			canvas.renderAll();
		}
	}
</script>

<div class="flex flex-col h-full bg-white rounded-xl border overflow-hidden">
	<!-- Toolbar -->
	<div class="flex items-center gap-2 px-4 py-2 border-b bg-gray-50 flex-wrap">
		<div class="flex gap-1">
			{#each [['pen', 'قلم'], ['eraser'], ['rect', 'مستطیل'], ['circle', 'دایره'], ['line', 'خط'], ['text', 'متن']] as [tool, label]}
				<button
					onclick={() => {
						if (['rect', 'circle', 'line', 'text'].includes(tool)) addShape(tool);
						else setTool(tool);
					}}
					class="px-3 py-1.5 text-xs rounded-lg transition-colors {activeTool === tool ? 'bg-blue-100 text-blue-700' : 'text-gray-600 hover:bg-gray-100'}"
					title={label || tool}
				>
					{label || tool === 'pen' ? 'قلم' : 'پاک‌کن'}
				</button>
			{/each}
		</div>

		<div class="w-px h-6 bg-gray-200"></div>

		<!-- Colors -->
		<div class="flex gap-1">
			{#each colors as color}
				<button
					class="w-6 h-6 rounded-full border-2 {brushColor === color ? 'border-blue-500 scale-110' : 'border-gray-200'}"
					style="background-color: {color}"
					onclick={() => setColor(color)}
				></button>
			{/each}
		</div>

		<div class="w-px h-6 bg-gray-200"></div>

		<!-- Width -->
		<input type="range" min="1" max="20" bind:value={brushWidth} onchange={() => setWidth(brushWidth)} class="w-20" />

		<div class="flex-1"></div>

		<button onclick={undo} class="px-3 py-1.5 text-xs text-gray-600 hover:bg-gray-100 rounded-lg">برگشت</button>
		<button onclick={clearCanvas} class="px-3 py-1.5 text-xs text-red-600 hover:bg-red-50 rounded-lg">پاک کردن</button>
	</div>

	<!-- Canvas -->
	<div class="flex-1 overflow-auto p-4 flex items-center justify-center bg-gray-100">
		<canvas bind:this={canvasEl}></canvas>
	</div>
</div>
