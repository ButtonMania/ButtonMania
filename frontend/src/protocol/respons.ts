import {ButtonPhase, GameState} from "./enums";

export class GameplayMessage {
    constructor(public readonly gameState: GameState,
                public readonly context?: SessionContext,
                public readonly record?: Record,
                public readonly error?: Error,
                public readonly bestUsersPayloads?: string[],
                public readonly countActive?: number,
                public readonly placeActive?: number,
                public readonly countLeaderboard?: number,
                public readonly bestOverallDuration?: number,
                public readonly bestTodaysDuration?: number,
                public readonly worldRecord?: boolean,
                public readonly message?: string,
                public readonly chat?: ChatMessage,
    ) {
    }
}

export class SessionContext {
    constructor(public readonly buttonPhase: ButtonPhase,
                public readonly timestamp: number,
                public readonly duration: number, 
                public readonly chat?: ChatMessage) { }
}

export class UpdateResponse {
    constructor(public readonly context: SessionContext,
                public readonly placeActive: number,
                public readonly countActive: number,
                public readonly message?: string) { }
}

export class Record {
    constructor(public readonly timestamp: number,
                public readonly duration: number) { }
}

export class RecordResponse {
    constructor(public readonly record: Record,
                public readonly placeLeaderboard: number,
                public readonly countLeaderboard: number,
                public readonly worldRecord: boolean) { }
}

export class Error {
    constructor(public readonly message: string) { }
}

export class ErrorResponse {
    constructor(public readonly error: Error) { }
}

export class RoomStats {
    constructor(public readonly countActive: number,
                public readonly countLeaderboard: number,
                public readonly bestOverallDuration: number,
                public readonly bestTodaysDuration: number,
                public readonly bestUsersPayloads: string[]) { }
}

export class ChatMessage {
    constructor(public readonly message: string, 
                public readonly userID?: string) { }
}
