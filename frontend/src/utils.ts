export function secondsToHHMMSS(seconds: number): string {
	const format: any = (value: number) => (value < 10 ? `0${value}` : value);
	var hours: number = Math.floor(seconds / 3600);
	var minutes: number = Math.floor((seconds - (hours * 3600)) / 60);
	seconds %= 60;
	return `${format(hours)}:${format(minutes)}:${format(seconds)}`;
}

export function randomEnum<T>(anEnum: T, minIndex?: number, maxIndex?: number): T[keyof T] {
	const enumValues = (Object.values(anEnum) as unknown) as T[keyof T][];
	const startIndex = minIndex ?? 0;
	const endIndex = maxIndex ?? enumValues.length;
	const randomIndex = Math.round(startIndex + (Math.random() * (endIndex - startIndex)));
	return enumValues[randomIndex];
}

export function enumIndex<T>(anEnum: T, enumValue: T[keyof T]): number {
	const enumValues = (Object.values(anEnum) as unknown) as T[keyof T][];
	return enumValues.indexOf(enumValue);
}