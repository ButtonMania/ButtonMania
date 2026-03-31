import { MessageBus } from '../src/services/bus';

class TestEvent {
    constructor(public readonly value: string) {}
}

class AnotherEvent {
    constructor(public readonly num: number) {}
}

describe('MessageBus', () => {
    beforeEach(() => {
        MessageBus.flush();
    });

    it('should deliver events to subscribers', async () => {
        let received: string | null = null;
        MessageBus.default(TestEvent).subscribe(async (event: TestEvent) => {
            received = event.value;
        });
        await MessageBus.raise(new TestEvent('hello'));
        expect(received).toBe('hello');
    });

    it('should support multiple subscribers', async () => {
        const results: string[] = [];
        MessageBus.default(TestEvent).subscribe(async (event: TestEvent) => {
            results.push('sub1:' + event.value);
        });
        MessageBus.default(TestEvent).subscribe(async (event: TestEvent) => {
            results.push('sub2:' + event.value);
        });
        await MessageBus.raise(new TestEvent('test'));
        expect(results).toEqual(['sub1:test', 'sub2:test']);
    });

    it('should isolate different event types', async () => {
        let testReceived = false;
        let anotherReceived = false;
        MessageBus.default(TestEvent).subscribe(async () => { testReceived = true; });
        MessageBus.default(AnotherEvent).subscribe(async () => { anotherReceived = true; });

        await MessageBus.raise(new TestEvent('x'));
        expect(testReceived).toBe(true);
        expect(anotherReceived).toBe(false);
    });

    it('should support unsubscribe', async () => {
        let count = 0;
        const handler = async () => { count++; };
        MessageBus.default(TestEvent).subscribe(handler);
        await MessageBus.raise(new TestEvent('1'));
        expect(count).toBe(1);

        MessageBus.default(TestEvent).unsubscribe(handler);
        await MessageBus.raise(new TestEvent('2'));
        expect(count).toBe(1);
    });

    it('should flush all channels', async () => {
        let received = false;
        MessageBus.default(TestEvent).subscribe(async () => { received = true; });
        MessageBus.flush();
        await MessageBus.raise(new TestEvent('x'));
        expect(received).toBe(false);
    });
});
