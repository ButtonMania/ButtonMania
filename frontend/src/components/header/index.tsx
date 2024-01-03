import { AppContext, AppContextProps } from '../context';
import { ButtonPhase, ButtonType } from '../../protocol/enums';
import { Component, h } from 'preact';

import { secondsToHHMMSS } from '../../utils';
import style from './style.css';

export default class Header extends Component {
    static contextType = AppContext;

    buttonStyleClass(buttonType: ButtonType): string {
        switch (buttonType) {
            case ButtonType.NewYear: {
                return style.newyear;
            }
            case ButtonType.Love: {
                return style.love;
            }
            case ButtonType.Fortune: {
                return style.fortune;
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
        let context: AppContextProps = this.context;
        let headerClasses: string[] = [
            style.header,
            this.buttonStyleClass(context.buttonType),
            this.buttonPhaseClass(context.buttonPhase)
        ];
        return (
            <header class={headerClasses.join(' ')}>
                <h1 class={style.type}>{context.headerText}</h1>
                <div class={style.message_wrapper}>
                    <div class={style.message_inner}>
                        <div class={style.record_wrapper}>
                            <span class={style.label}>{context.bestTodaysDurationText}</span>
                            <span class={style.value}>{secondsToHHMMSS(context.bestTodaysDurationValue)}</span>
                        </div>
                        <h6 class={style.message}>{context.messageText}</h6>
                        <div class={style.record_wrapper}>
                            <span class={style.label}>{context.bestOverallDurationText}</span>
                            <span class={style.value}>{secondsToHHMMSS(context.bestOverallDurationValue)}</span>
                        </div>
                    </div>
                </div>
            </header>
        );
    }
}