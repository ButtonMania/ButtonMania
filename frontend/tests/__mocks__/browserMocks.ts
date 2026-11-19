// Mock Browser API's which are not supported by JSDOM, e.g. ServiceWorker, LocalStorage
import { TextDecoder, TextEncoder } from 'util';

if (typeof (global as any).TextEncoder === 'undefined') {
	(global as any).TextEncoder = TextEncoder;
}
if (typeof (global as any).TextDecoder === 'undefined') {
	(global as any).TextDecoder = TextDecoder;
}
/**
 * An example how to mock localStorage is given below 👇
 */

/* 
// Mocks localStorage
const localStorageMock = (function() {
	let store = {};

	return {
		getItem: (key) => store[key] || null,
		setItem: (key, value) => store[key] = value.toString(),
		clear: () => store = {}
	};

})();

Object.defineProperty(window, 'localStorage', {
	value: localStorageMock
}); */
