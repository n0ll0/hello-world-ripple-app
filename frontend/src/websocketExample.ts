// Example: How to use WebSocket in your Ripple components
// Each event type has its own dedicated WebSocket connection

// import { effect } from 'ripple';
import { todoCreatedWS, todoUpdatedWS, todoDeletedWS } from '@/useWebSocket.ts';

// Example 1: Connect to specific event WebSockets when component mounts
export function setupWebSockets() {
	// Connect to each event-specific WebSocket
	todoCreatedWS.connect();
	todoUpdatedWS.connect();
	todoDeletedWS.connect();

	// Listen for todo created events
	const unsubCreated = todoCreatedWS.onMessage((todo) => {
		console.log('Todo created:', todo);
		// Handle new todo created by another user
	});

	// Listen for todo updated events
	const unsubUpdated = todoUpdatedWS.onMessage((todo) => {
		console.log('Todo updated:', todo);
		// Handle todo update from another user
	});

	// Listen for todo deleted events
	const unsubDeleted = todoDeletedWS.onMessage((data) => {
		console.log('Todo deleted:', data);
		// Handle todo deletion from another user
	});

	// Cleanup function (call when component unmounts)
	return () => {
		unsubCreated();
		unsubUpdated();
		unsubDeleted();
		todoCreatedWS.disconnect();
		todoUpdatedWS.disconnect();
		todoDeletedWS.disconnect();
	};
}

// Example 2: Integrate WebSocket with Ripple reactive system
export function useRealtimeTodos($todos) {
	// Connect to all todo event WebSockets
	todoCreatedWS.connect();
	todoUpdatedWS.connect();
	todoDeletedWS.connect();

	// Listen for created events
	const unsubCreated = todoCreatedWS.onMessage((todo) => {
		try {
			const data = typeof todo === 'string' ? JSON.parse(todo) : todo;
			// Add new todo to the tracked map
			$todos.set(data.id, data);
		} catch (e) {
			console.error('Error parsing created todo:', e);
		}
	});

	// Listen for updated events
	const unsubUpdated = todoUpdatedWS.onMessage((todo) => {
		try {
			const data = typeof todo === 'string' ? JSON.parse(todo) : todo;
			// Update existing todo
			if ($todos.has(data.id)) {
				$todos.set(data.id, data);
			}
		} catch (e) {
			console.error('Error parsing updated todo:', e);
		}
	});

	// Listen for deleted events
	const unsubDeleted = todoDeletedWS.onMessage((payload) => {
		try {
			const data = typeof payload === 'string' ? JSON.parse(payload) : payload;
			// Remove todo from tracked map
			$todos.delete(data.id);
		} catch (e) {
			console.error('Error parsing deleted todo:', e);
		}
	});

	// Return cleanup function
	return () => {
		unsubCreated();
		unsubUpdated();
		unsubDeleted();
		todoCreatedWS.disconnect();
		todoUpdatedWS.disconnect();
		todoDeletedWS.disconnect();
	};
}

// Example usage in a component:
/*
import { setupWebSockets, useRealtimeTodos } from '@/websocketExample.ts';
import { ripple } from '@/types.ts';

component MyTodoApp() {
	const $todos = new TrackedMap();
	
	// Set up WebSocket on mount for real-time updates
	effect(() => {
		const cleanup = useRealtimeTodos($todos);
		return cleanup; // This will be called when effect is cleaned up
	});
	
	// When you create a todo, it will automatically be broadcast to other clients
	async function addTodo(todo) {
		const {data: $todo} = await createTodo(todo);
		$todos.set($todo.id, ripple($todo));
		// All other connected clients will receive this via /ws/todos/created
	}
	
	// Your component JSX...
}
*/
