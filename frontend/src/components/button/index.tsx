import { ButtonPhase, ButtonType } from '../../protocol/enums';
import { Component, h } from 'preact';

import { AppContext } from '../context';
import ArrowLeft from './arrow_left.svg';
import ArrowRight from './arrow_right.svg';
import Circles from './circles.svg';
import CirclesRecord from './circles_record.svg';
import { Player } from "@lottiefiles/react-lottie-player";
import { secondsToHHMMSS } from '../../utils';
import style from './style.css';

interface Props {
    buttonHandler: (buttonPhase: ButtonPhase, holdDuration: number) => void;
    arrowsHandler: (buttonType: ButtonType) => void;
}

interface State {
    wakeLock?: WakeLockSentinel;
    intervalId?: number;
    timeoutId?: number;
    visibilityListener: any;
}

export default class Button extends Component<Props, State> {
    static contextType = AppContext;
    private readonly pushTimeout: number = 80;
    private readonly holdTimeout: number = 1000;

    componentDidMount(): void {
        // Listen to visibility change event, cancel hold if screen is locked
        let visibilityListener = this.onVisibilityChange.bind(this);
        document.addEventListener('visibilitychange', visibilityListener);
        document.addEventListener('webkitvisibilitychange', visibilityListener);
        this.setState({
            visibilityListener: visibilityListener,
        });
    }

    componentWillUnmount(): void {
        this.cancelHold();
        // Remove visibilitychange event listener
        document.removeEventListener('visibilitychange', this.state.visibilityListener);
        document.removeEventListener('webkitvisibilitychange', this.state.visibilityListener);
    }

    async requestScreenLock(): Promise<WakeLockSentinel> {
        try {
            return await navigator.wakeLock.request();
        } catch (err) {
            console.error(`${err.name}, ${err.message}`);
        }
    }

    async cancelHold(): Promise<void> {
        return new Promise((resolve) => {
            // Release wake lock
            this.state.wakeLock?.release();
            // Cancel interval and timers
            if (this.state.intervalId != null) {
                window.clearInterval(this.state.intervalId);
            }
            if (this.state.timeoutId != null) {
                window.clearTimeout(this.state.timeoutId);
            }
            this.setState({
                intervalId: null,
                timeoutId: null,
                wakeLock: null,
            }, () => { resolve() });
        })
    }

    async startHold(): Promise<void> {
        return new Promise((resolve) => {
            // Prevent screen to sleep
            this.requestScreenLock().then((wakeLock: WakeLockSentinel) => {
                let buttonPhase = this.context.buttonPhase;
                var timeoutId = this.state.timeoutId;
                if (buttonPhase == ButtonPhase.Idle || buttonPhase == ButtonPhase.Release) {
                    // Change phase to hold after little delay
                    timeoutId = window.setTimeout(
                        this.onAfterPush.bind(this),
                        this.pushTimeout
                    );
                }
                this.setState({
                    timeoutId: timeoutId,
                    wakeLock: wakeLock,
                }, () => { resolve() });
            });
        })
    }

    holdTimeInSec(): number {
        const nowTs: number = Math.floor(Date.now() / 1000) | 0;
        const hold: number = nowTs - this.context.pushTimestamp;
        return hold > 0 ? hold : 0;
    }

    onRelease(): void {
        this.cancelHold().then(() => {
            const hold: number = this.holdTimeInSec();
            if (this.context.buttonPhase != ButtonPhase.Idle) {
                this.props.buttonHandler(ButtonPhase.Release, hold);
            }
        });
    }

    onHold(): void {
        const hold: number = this.holdTimeInSec();
        this.props.buttonHandler(ButtonPhase.Hold, hold);
    }

    onAfterPush(): void {
        this.setState({
            intervalId: window.setInterval(this.onHold.bind(this), this.holdTimeout),
        }, () => { this.props.buttonHandler(ButtonPhase.Hold, 0) });
    }

    onPush(): void {
        this.startHold().then(() => {
            this.props.buttonHandler(ButtonPhase.Push, 0);
        });
    }

    onTouchStart(e: TouchEvent): void {
        this.onPush();
        e.preventDefault();
    }

    onTouchEnd(e: TouchEvent): void {
        this.onRelease();
        e.preventDefault();
    }

    onTouchCancel(e: TouchEvent): void {
        this.onRelease();
        e.preventDefault();
    }

    onMouseDown(e: MouseEvent): void {
        if (e.button == 0) {
            this.onPush();
        }
        e.preventDefault();
    }

    onMouseUp(e: MouseEvent): void {
        this.onRelease();
        e.preventDefault();
    }

    onMouseLeave(e: MouseEvent): void {
        let buttonPhase = this.context.buttonPhase;
        if (buttonPhase == ButtonPhase.Hold || buttonPhase == ButtonPhase.Push) {
            this.onRelease();
        }
        e.preventDefault();
    }

    onVisibilityChange(e: Event): void {
        this.onRelease();
        e.preventDefault();
    }

    onClickLeftArrow(e: MouseEvent): void {
        let prevButtonType: ButtonType = this.prevButtonType(
            this.context.buttonType,
            this.context.isPremium
        );
        this.props.arrowsHandler(prevButtonType);
        e.preventDefault();
    }

    onClickRightArrow(e: MouseEvent): void {
        let nextButtonType: ButtonType = this.nextButtonType(
            this.context.buttonType,
            this.context.isPremium
        );
        this.props.arrowsHandler(nextButtonType);
        e.preventDefault();
    }

    nextButtonType(buttonType: ButtonType, isPremium: boolean): ButtonType {
        switch (buttonType) {
            case ButtonType.Love: {
                return ButtonType.Fortune;
            }
            case ButtonType.Fortune: {
                return ButtonType.Peace
            }
            case ButtonType.Peace: {
                return isPremium ? ButtonType.Prestige : ButtonType.Love;
            }
            case ButtonType.Prestige: {
                return ButtonType.Love;
            }
        }
    }

    prevButtonType(buttonType: ButtonType, isPremium: boolean): ButtonType {
        switch (buttonType) {
            case ButtonType.Love: {
                return isPremium ? ButtonType.Prestige : ButtonType.Peace;
            }
            case ButtonType.Fortune: {
                return ButtonType.Love;
            }
            case ButtonType.Peace: {
                return ButtonType.Fortune;
            }
            case ButtonType.Prestige: {
                return ButtonType.Peace;
            }
        }
    }

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
            case ButtonPhase.Push: {
                return style.push;
            }
            case ButtonPhase.Hold: {
                return style.hold;
            }
            case ButtonPhase.Release: {
                return style.release;
            }
        }
    }

    render(props: Props, state: State) {
        let worldRecord: boolean = this.context.worldRecord;
        let buttonClasses: string[] = [
            style.button,
            this.buttonPhaseClass(this.context.buttonPhase),
            this.buttonStyleClass(this.context.buttonType)
        ];
        if (worldRecord) {
            buttonClasses.push(style.world_record);
        }
        return (
            <section class={buttonClasses.join(' ')}>
                <div class={style.wrapper}>
                    <div class={style.inner}>
                        <div class={style.labels_wrapper}>
                            <span class={style.time}>{secondsToHHMMSS(this.context.holdDuration)}</span>
                            <span class={style.label}>{this.context.buttonText}</span>
                        </div>
                        <div class={style.animation}>
                            <Player
                                autoplay
                                loop
                                src={this.context.buttonAnimation}
                            />
                        </div>
                    </div>
                </div>
                <div class={style.arrows}>
                    <a href='#' onClick={this.onClickLeftArrow.bind(this)}><ArrowLeft /></a>
                    <a href='#' onClick={this.onClickRightArrow.bind(this)}><ArrowRight /></a>
                </div>
                <div
                    class={style.gradient}
                    onMouseDown={this.onMouseDown.bind(this)}
                    onMouseUp={this.onMouseUp.bind(this)}
                    onMouseLeave={this.onMouseLeave.bind(this)}
                    onTouchStart={this.onTouchStart.bind(this)}
                    onTouchEnd={this.onTouchEnd.bind(this)}
                    onTouchCancel={this.onTouchCancel.bind(this)}
                />
                <div class={style.circles}>
                    {!worldRecord && <Circles />}
                    {worldRecord && <CirclesRecord />}
                </div>
            </section >
        );
    }
}