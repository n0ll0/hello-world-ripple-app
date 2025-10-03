# WebSocket Event Architecture

This application uses a **one event per endpoint** WebSocket architecture. Each type of event has its own dedicated WebSocket endpoint.

## Backend Endpoints

The backend provides three WebSocket endpoints for todo events:

- `ws://localhost:8080/ws/todos/created` - Broadcasts when a todo is created
- `ws://localhost:8080/ws/todos/updated` - Broadcasts when a todo is updated
- `ws://localhost:8080/ws/todos/deleted` - Broadcasts when a todo is deleted

### How it Works

1. **EventHub**: Each endpoint has its own `EventHub` that manages WebSocket connections for that specific event type
2. **Auto-broadcast**: When a todo is created, updated, or deleted via the REST API, the event is automatically broadcast to all clients subscribed to that event's WebSocket
3. **JSON Messages**: Events are sent as JSON-encoded messages

#### Event Payloads

**Created/Updated Events:**
```json
{
  "id": 1,
  "user_id": 1,
  "title": "Buy groceries",
  "completed": false,
  "created_at": "2025-10-03T14:30:00Z",
  "updated_at": "2025-10-03T14:30:00Z"
}
```

**Deleted Events:**
```json
{
  "id": 1
}
```

## Frontend Usage

### WebSocket Clients

Import the pre-configured WebSocket clients:

```typescript
import { todoCreatedWS, todoUpdatedWS, todoDeletedWS } from '@/useWebSocket.ts';
```

### Basic Usage

```typescript
// Connect to specific events
todoCreatedWS.connect();
todoUpdatedWS.connect();
todoDeletedWS.connect();

// Listen for messages
todoCreatedWS.onMessage((todo) => {
  console.log('New todo created:', todo);
});

todoUpdatedWS.onMessage((todo) => {
  console.log('Todo updated:', todo);
});

todoDeletedWS.onMessage((data) => {
  console.log('Todo deleted:', data.id);
});

// Cleanup
todoCreatedWS.disconnect();
```

### Integration with Ripple

For reactive real-time updates with your Ripple `TrackedMap`:

```typescript
import { effect, TrackedMap } from 'ripple';
import { ripple } from '@/types.ts';
import { useRealtimeTodos } from '@/websocketExample.ts';

component TodoApp() {
  const $todos = new TrackedMap();
  
  // Set up real-time WebSocket updates
  effect(() => {
    const cleanup = useRealtimeTodos($todos);
    return cleanup; // Auto-cleanup on unmount
  });
  
  // Your todos will now update in real-time across all clients!
  // When any client creates/updates/deletes a todo, all other clients see it instantly
}
```

## Benefits of This Architecture

1. **Separation of Concerns**: Each event type has its own connection, making it easier to debug and monitor
2. **Selective Subscriptions**: Clients can subscribe only to the events they care about
3. **Cleaner Code**: No need for message type parsing or routing on the client side
4. **Better Performance**: Clients only receive messages for events they're subscribed to
5. **Scalability**: Each event hub can be scaled independently if needed

## Example: Full Real-Time Todo App

```typescript
import { effect, TrackedMap } from 'ripple';
import { ripple, unripple } from '@/types.ts';
import { createTodo, editTodo, deleteTodo, getTodos } from '@/useApi.ts';
import { todoCreatedWS, todoUpdatedWS, todoDeletedWS } from '@/useWebSocket.ts';

component TodoApp() {
  const $todos = new TrackedMap();
  
  // Load initial todos
  effect(async () => {
    const { data: todos } = await getTodos();
    if (todos) {
      for (const todo of todos) {
        $todos.set(todo.id, ripple(todo));
      }
    }
  });
  
  // Set up real-time updates
  effect(() => {
    todoCreatedWS.connect();
    todoUpdatedWS.connect();
    todoDeletedWS.connect();
    
    const unsubCreated = todoCreatedWS.onMessage((todo) => {
      const data = typeof todo === 'string' ? JSON.parse(todo) : todo;
      if (!$todos.has(data.id)) {
        $todos.set(data.id, ripple(data));
      }
    });
    
    const unsubUpdated = todoUpdatedWS.onMessage((todo) => {
      const data = typeof todo === 'string' ? JSON.parse(todo) : todo;
      $todos.set(data.id, ripple(data));
    });
    
    const unsubDeleted = todoDeletedWS.onMessage((payload) => {
      const data = typeof payload === 'string' ? JSON.parse(payload) : payload;
      $todos.delete(data.id);
    });
    
    return () => {
      unsubCreated();
      unsubUpdated();
      unsubDeleted();
      todoCreatedWS.disconnect();
      todoUpdatedWS.disconnect();
      todoDeletedWS.disconnect();
    };
  });
  
  // Create a new todo (will be broadcast to all clients)
  async function addTodo(title: string) {
    const { data } = await createTodo({ title });
    // No need to manually update $todos - the WebSocket will handle it!
  }
  
  // Your component JSX...
}
```

## Testing WebSockets

You can test WebSocket connections using:

1. **Browser DevTools**: Open Network tab, filter by WS
2. **wscat**: `wscat -c ws://localhost:8080/ws/todos/created`
3. **Postman**: Supports WebSocket connections

Example with wscat:
```bash
# Terminal 1: Connect to created events
wscat -c ws://localhost:8080/ws/todos/created

# Terminal 2: Create a todo via REST API
curl -X POST http://localhost:8080/api/todos \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title": "Test Todo"}'

# Terminal 1 will receive the broadcast!
```
