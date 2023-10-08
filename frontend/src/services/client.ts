import { ButtonType, MessageType, UserLocale } from "../protocol/enums";
import { ConnectEvent, DisconnectEvent } from "../protocol/internal";
import { ErrorResponse, RecordResponse, RoomStats, UpdateResponse } from "../protocol/respons";

import { Config } from "../config";
import { GameplayMessageRequest } from "../protocol/requests";
import { MessageBus } from "./bus";
import axios from 'axios';
import { plainToInstance } from 'class-transformer';

export class NetworkClient {
    private ws?: WebSocket;
    private wsEndpoint: string;
    private httpEndpoint: string;

    constructor(public config: Config) {
        this.wsEndpoint = config.wsEndpoint;
        this.httpEndpoint = config.httpEndpoint;
        MessageBus.default(GameplayMessageRequest).subscribe(this.sendMessage.bind(this));
    }

    private establishWsConnection(url: URL): WebSocket {
        let ws = new WebSocket(url);
        ws.binaryType = 'arraybuffer';

        ws.onopen = this.onWsOpen.bind(this);
        ws.onmessage = this.onWsMessage.bind(this);
        ws.onerror = this.onWsError.bind(this);
        ws.onclose = this.onWsClose.bind(this);
        return ws;
    }

    private closeWsConnection(): void {
        if (this.ws != null && this.ws.readyState !== WebSocket.CLOSED) {
            this.ws!.close();
            this.ws = null;
        }
    }

    private onWsOpen(_: Event): void {
        MessageBus.raise(new ConnectEvent()).catch(console.log);
    }

    private onWsError(_: Event): void {
        MessageBus.raise(new DisconnectEvent()).catch(console.log);
    }

    private onWsClose(ev: CloseEvent): void {
        MessageBus.raise(new DisconnectEvent()).catch(console.log);
    }

    private onWsMessage(ev: MessageEvent): void {
        let raw: any = JSON.parse(ev.data);

        switch (raw.messageType) {
            case MessageType.Update: {
                MessageBus.raise(plainToInstance(UpdateResponse, raw as object)).catch(console.log);
                break;
            }
            case MessageType.Record: {
                MessageBus.raise(plainToInstance(RecordResponse, raw as object)).catch(console.log);
                break;
            }
            case MessageType.Error: {
                MessageBus.raise(plainToInstance(ErrorResponse, raw as object)).catch(console.log);
                break;
            }
        }
    }

    private sendMessage(message: GameplayMessageRequest): void {
        try {
            if (this.ws != null && this.ws.readyState === WebSocket.OPEN) {
                this.ws!.send(JSON.stringify(message));
            }
        }
        catch (e) {
            console.log(e)
        }
    }

    public fetchRoomStats(buttonType: ButtonType): void {
        axios.get(`${this.httpEndpoint}/stats`, {
            params: { buttonType: buttonType }
        }).then(response => {
            let raw: any = response.data as object;
            let stats: RoomStats = plainToInstance(RoomStats, raw as object);
            MessageBus.raise(stats).catch(console.log);
        });
    }

    public startGameplay(initData: string, buttonType: ButtonType): void {
        this.closeWsConnection();
        var url = new URL(this.wsEndpoint);
        url.searchParams.append('initData', initData);
        url.searchParams.append('buttonType', buttonType);
        this.ws = this.establishWsConnection(url);
    }

    public stopGameplay(): void {
        this.closeWsConnection();
    }
}