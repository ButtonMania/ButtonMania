/// <reference types="preact/compat" />
type WakeLockType = "screen";
interface WakeLockSentinel extends EventTarget {
  released: boolean;
  release(): Promise<void>;
  onrelease: ((this: WakeLockSentinel, ev: Event) => any) | null;
}
interface WakeLock {
  request(type?: WakeLockType): Promise<WakeLockSentinel>;
}
interface Navigator {
  wakeLock?: WakeLock;
}
