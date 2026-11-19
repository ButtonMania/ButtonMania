import { ButtonPhase } from "./enums";

export class GameplayMessageRequest {
    public readonly buttonPhase: ButtonPhase;
    constructor(buttonPhase: ButtonPhase) {
        this.buttonPhase = buttonPhase;
    }
}