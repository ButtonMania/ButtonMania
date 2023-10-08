interface IMessageReceiver<T> {
    (data: T): Promise<void>;
}

function messageBusCreator<T>() {
    class Channel {
        private static instance: Channel = new Channel();
        private subscribers: IMessageReceiver<T>[] = [];

        public static get default(): Channel {
            return Channel.instance;
        }

        public static get new(): Channel {
            return new Channel();
        }

        public subscribe(receiver: IMessageReceiver<T>): void {
            this.subscribers.push(receiver);
        }

        public unsubscribe(receiver: IMessageReceiver<T>): void {
            let idx: number;
            while ((idx = this.subscribers.indexOf(receiver)) >= 0) {
                this.subscribers.splice(idx, 1);
            }
        }

        public async raise(data: T): Promise<void> {
            for (let subscriber of this.subscribers) {
                await subscriber(data);
            }
        }

        public flush(): void {
            this.subscribers = [];
        }
    }

    return Channel;
}

export class MessageBus {
    private static _typeCache: Map<any, Function> = new Map<any, Function>();

    private static get typeCache(): Map<any, Function> {
        if (!MessageBus._typeCache) {
            MessageBus._typeCache = new Map<any, Function>();
        }
        return MessageBus._typeCache;
    }

    public static typeProvider<T>(key: any): any {
        if (MessageBus.typeCache.has(key)) {
            return MessageBus.typeCache.get(key);
        }

        let instance = messageBusCreator<T>();
        MessageBus.typeCache.set(key, instance);
        return instance;
    }

    public static default<T>(key: any) {
        return MessageBus.typeProvider<T>(key).default;
    }

    public static new<T>(key: any) {
        return MessageBus.typeProvider<T>(key).new;
    }

    public static async raise<T extends object>(event: T) {
        return await MessageBus.default<T>(event.constructor).raise(event);
    }

    public static flush() {
        MessageBus._typeCache = new Map<any, Function>();
    }
}
