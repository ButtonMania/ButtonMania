import { ButtonType, GameState } from "../protocol/enums";
import { ConnectEvent, DisconnectEvent } from "../protocol/internal";
import {ErrorResponse, GameplayMessage, RecordResponse, RoomStats, UpdateResponse} from "../protocol/respons";

import { Config } from "../config";
import { GameplayMessageRequest } from "../protocol/requests";
import { MessageBus } from "./bus";
import axios from 'axios';
import { plainToInstance } from 'class-transformer';

export class NetworkClient {
    private ws?: WebSocket;
    private wsEndpoint: string;
    private httpEndpoint: string;
    private clientId: string;

    constructor(public config: Config) {
        this.wsEndpoint = config.wsEndpoint;
        this.httpEndpoint = config.httpEndpoint;
        this.clientId = config.clientId;
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

    private onWsClose(_: CloseEvent): void {
        MessageBus.raise(new DisconnectEvent()).catch(console.log);
    }

    private onWsMessage(ev: MessageEvent): void {
        let raw: GameplayMessage = JSON.parse(ev.data);

        switch (raw.gameState) {
            case GameState.Update: {
                MessageBus.raise(plainToInstance(UpdateResponse, raw as object)).catch(console.log);
                break;
            }
            case GameState.Record: {
                MessageBus.raise(plainToInstance(RecordResponse, raw as object)).catch(console.log);
                break;
            }
            case GameState.Error: {
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
        axios.get(`${this.httpEndpoint}/room/stats`, {
            params: { clientId: this.clientId, roomId: buttonType }
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
        url.searchParams.append('roomId', buttonType);
        url.searchParams.append('clientId', this.clientId);
        this.ws = this.establishWsConnection(url);
    }

    public stopGameplay(): void {
        this.closeWsConnection();
    }
}