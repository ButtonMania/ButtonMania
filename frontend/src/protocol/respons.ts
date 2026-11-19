import {ButtonPhase, GameState} from "./enums";

export class GameplayMessage {
    public readonly gameState: GameState;
    public readonly context?: SessionContext;
    public readonly record?: Record;
    public readonly error?: Error;
    public readonly bestUsersPayloads?: string[];
    public readonly countActive?: number;
    public readonly placeActive?: number;
    public readonly countLeaderboard?: number;
    public readonly bestOverallDuration?: number;
    public readonly bestTodaysDuration?: number;
    public readonly worldRecord?: boolean;
    public readonly message?: string;
    public readonly chat?: ChatMessage;

    constructor(gameState: GameState,
                context?: SessionContext,
                record?: Record,
                error?: Error,
                bestUsersPayloads?: string[],
                countActive?: number,
                placeActive?: number,
                countLeaderboard?: number,
                bestOverallDuration?: number,
                bestTodaysDuration?: number,
                worldRecord?: boolean,
                message?: string,
                chat?: ChatMessage,
    ) {
        this.gameState = gameState;
        this.context = context;
        this.record = record;
        this.error = error;
        this.bestUsersPayloads = bestUsersPayloads;
        this.countActive = countActive;
        this.placeActive = placeActive;
        this.countLeaderboard = countLeaderboard;
        this.bestOverallDuration = bestOverallDuration;
        this.bestTodaysDuration = bestTodaysDuration;
        this.worldRecord = worldRecord;
        this.message = message;
        this.chat = chat;
    }
}

export class SessionContext {
    public readonly buttonPhase: ButtonPhase;
    public readonly timestamp: number;
    public readonly duration: number;
    public readonly chat?: ChatMessage;

    constructor(buttonPhase: ButtonPhase,
                timestamp: number,
                duration: number,
                chat?: ChatMessage) {
        this.buttonPhase = buttonPhase;
        this.timestamp = timestamp;
        this.duration = duration;
        this.chat = chat;
    }
}

export class UpdateResponse {
    public readonly context: SessionContext;
    public readonly placeActive: number;
    public readonly countActive: number;
    public readonly message?: string;

    constructor(context: SessionContext,
                placeActive: number,
                countActive: number,
                message?: string) {
        this.context = context;
        this.placeActive = placeActive;
        this.countActive = countActive;
        this.message = message;
    }
}

export class Record {
    public readonly timestamp: number;
    public readonly duration: number;

    constructor(timestamp: number,
                duration: number) {
        this.timestamp = timestamp;
        this.duration = duration;
    }
}

export class RecordResponse {
    public readonly record: Record;
    public readonly placeLeaderboard: number;
    public readonly countLeaderboard: number;
    public readonly worldRecord: boolean;

    constructor(record: Record,
                placeLeaderboard: number,
                countLeaderboard: number,
                worldRecord: boolean) {
        this.record = record;
        this.placeLeaderboard = placeLeaderboard;
        this.countLeaderboard = countLeaderboard;
        this.worldRecord = worldRecord;
    }
}

export class Error {
    public readonly message: string;

    constructor(message: string) {
        this.message = message;
    }
}

export class ErrorResponse {
    public readonly error: Error;

    constructor(error: Error) {
        this.error = error;
    }
}

export class RoomStats {
    public readonly countActive: number;
    public readonly countLeaderboard: number;
    public readonly bestOverallDuration: number;
    public readonly bestTodaysDuration: number;
    public readonly bestUsersPayloads: string[];

    constructor(countActive: number,
                countLeaderboard: number,
                bestOverallDuration: number,
                bestTodaysDuration: number,
                bestUsersPayloads: string[]) {
        this.countActive = countActive;
        this.countLeaderboard = countLeaderboard;
        this.bestOverallDuration = bestOverallDuration;
        this.bestTodaysDuration = bestTodaysDuration;
        this.bestUsersPayloads = bestUsersPayloads;
    }
}

export class ChatMessage {
    public readonly message: string;
    public readonly userID?: string;

    constructor(message: string,
                userID?: string) {
        this.message = message;
        this.userID = userID;
    }
}
