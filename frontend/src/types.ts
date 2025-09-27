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

export type Todo = {};
export type r_Todo = ripple<Todo>;

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
