import { ButtonPhase } from "./enums";

export class GameplayMessageRequest {
    constructor(public readonly buttonPhase: ButtonPhase) {
    }
}