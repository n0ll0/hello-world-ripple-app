import { Todo, User } from "@/types.ts";

interface api_response<T> {
	data: T;
	error?: Error;
}

const API_BASE = (import.meta.env?.VITE_API_BASE_URL as string | undefined) ?? "http://localhost:8080";

async function request<T>(path: string, init: RequestInit = {}): Promise<api_response<T>> {
	const url = new URL(path, API_BASE).toString();
	const headers = new Headers(init.headers ?? {});

	const hasBody = init.body !== undefined && !(init.body instanceof FormData);
	if (hasBody && !headers.has("Content-Type")) {
		headers.set("Content-Type", "application/json");
	}

	const token = typeof window !== "undefined" ? window.localStorage.getItem("access_token") : null;
	if (token && !headers.has("Authorization")) {
		headers.set("Authorization", `Bearer ${token}`);
	}

	try {
		const response = await fetch(url, { ...init, headers });
		if (!response.ok) {
			const message = await response.text();
			return {
				data: undefined as unknown as T,
				error: new Error(message || response.statusText),
			};
		}

		let payload: T | undefined;
		if (response.status !== 204) {
			const text = await response.text();
			payload = text ? (JSON.parse(text) as T) : (undefined as unknown as T);
		}

		return { data: payload ?? (undefined as unknown as T) };
	} catch (err) {
		console.error(err);
		return {
			data: undefined as unknown as T,
			error: err instanceof Error ? err : new Error("Unknown error"),
		};
	}
}

export function createTodo(todo: Todo): Promise<api_response<Todo>> {
	return request<Todo>("/api/todos", {
		method: "POST",
		body: JSON.stringify(todo)
	});
}

export function getTodos(): Promise<api_response<Todo[]>> {
	return request<Todo[]>("/api/todos");
}

export function getTodo(id: string): Promise<api_response<Todo>> {
	return request<Todo>(`/api/todos/${encodeURIComponent(id)}`);
}

export function editTodo(
	id: string,
	update: Partial<Pick<Todo, "title" | "completed">> = {},
): Promise<api_response<Todo>> {
	return request<Todo>(`/api/todos/${encodeURIComponent(id)}`, {
		method: "PUT",
		body: JSON.stringify(update),
	});
}

export function deleteTodo(id: string): Promise<api_response<void>> {
	return request<void>(`/api/todos/${encodeURIComponent(id)}`, {
		method: "DELETE",
	});
}

export async function login(credentials: { username: string; password: string }): Promise<api_response<{ access_token: string }>> {
	const params = new URLSearchParams();
	params.set("grant_type", "password");
	params.set("username", credentials.username);
	params.set("password", credentials.password);
	// Use your OAuth2 client credentials here:
	// If you have only one client, you can hardcode or load from env
	const clientId = import.meta.env?.VITE_OAUTH_CLIENT_ID ?? "client-id";
	const clientSecret = import.meta.env?.VITE_OAUTH_CLIENT_SECRET ?? "client-secret";
	params.set("client_id", clientId);
	params.set("client_secret", clientSecret);

	const url = new URL("/token", API_BASE).toString();
	try {
		const response = await fetch(url, {
			method: "POST",
			headers: { "Content-Type": "application/x-www-form-urlencoded" },
			body: params.toString(),
		});
		const data = await response.json();
		if (response.ok && data.access_token) {
			if (typeof window !== "undefined") {
				window.localStorage.setItem("access_token", data.access_token);
			}
			return { data };
		} else {
			return { data: undefined as any, error: new Error(data.error_description || data.error || response.statusText) };
		}
	} catch (err) {
		return { data: undefined as any, error: err instanceof Error ? err : new Error("Unknown error") };
	}
}

export async function register(credentials: { username: string; password: string }): Promise<api_response<User>> {
	return request<User>("/api/users", {
		method: "POST",
		body: JSON.stringify(credentials),
	});
}

export function getCurrentUser(): Promise<api_response<User>> {
	return request<User>("/api/users/me");
}

