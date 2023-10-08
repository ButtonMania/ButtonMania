import { ButtonPhase, ButtonType } from '../../protocol/enums';
import { Component, h } from 'preact';

import { AppContext } from '../context';
import style from './style.css';

export default class Header extends Component {
    static contextType = AppContext;

    buttonStyleClass(buttonType: ButtonType): string {
        switch (buttonType) {
            case ButtonType.Fortune: {
                return style.fortune;
            }
            case ButtonType.Love: {
                return style.love;
            }
            case ButtonType.Peace: {
                return style.peace;
            }
            case ButtonType.Prestige: {
                return style.prestige;
            }
        }
    }

    buttonPhaseClass(buttonPhase: ButtonPhase): string {
        switch (buttonPhase) {
            case ButtonPhase.Idle: {
                return style.idle;
            }
            case ButtonPhase.Push:
            case ButtonPhase.Hold: {
                return style.hold;
            }
            case ButtonPhase.Release: {
                return style.normal;
            }
        }
    }

    render() {
        let headerClasses: string[] = [
            style.header,
            this.buttonStyleClass(this.context.buttonType),
            this.buttonPhaseClass(this.context.buttonPhase)
        ];
        return (
            <header class={headerClasses.join(' ')}>
                <h1 class={style.type}>{this.context.headerText}</h1>
                <h6 class={style.message}>{this.context.messageText}</h6>
            </header>
        );
    }
}