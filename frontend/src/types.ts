// type ripple_value = string | number | boolean;
type ripple<T> = T extends object
	? {
		[K in keyof T as K extends `$${string}`
		? K
		: K extends `${infer Name}`
		? `$${Name}`
		: K]: T[K] extends object ? ripple<T[K]> : T[K];
	}
	: T;

type unripple<T> = T extends object
	? {
		[K in keyof T as K extends `$${infer Name}`
		? Name
		: K]: T[K] extends object ? unripple<T[K]> : T[K];
	}
	: T;

export type Todo = {
	id: number;
	userId: number;
	title: string;
	completed: boolean;
	createdAt: string;
};
export type r_Todo = ripple<Todo>;

export type User = {
	id: number;
	username: string;
	createdAt: string;
};
export type r_User = ripple<User>;

export function ripple<T>(value: T): ripple<T> {
	if (typeof value !== 'object' || value === null) {
		return value as ripple<T>;
	}
	const result: any = Array.isArray(value) ? [] : {};
	for (const key in value) {
		if (Object.prototype.hasOwnProperty.call(value, key)) {
			const newKey = key.startsWith('$') ? key : `$${key}`;
			const val = (value as any)[key];
			result[newKey] = typeof val === 'object' ? ripple(val) : val;
		}
	}
	return result;
}


export type u_Todo = unripple<r_Todo>;

export function unripple<T>(value: T): unripple<T> {
	if (typeof value !== 'object' || value === null) {
		return value as unripple<T>;
	}
	const result: any = Array.isArray(value) ? [] : {};
	for (const key in value) {
		if (Object.prototype.hasOwnProperty.call(value, key)) {
			const newKey = key.startsWith('$') ? key.slice(1) : key;
			const val = (value as any)[key];
			result[newKey] = typeof val === 'object' ? unripple(val) : val;
		}
	}
	return result;
}
