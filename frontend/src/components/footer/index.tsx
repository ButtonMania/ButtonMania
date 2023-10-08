import { Component, h } from 'preact';

import { AppContext } from '../context';
import { ButtonPhase } from '../../protocol/enums';
import Counter from '../counter';
import style from './style.css';

interface Props {
    placeActiveText: string;
    placeLeaderboardText: string;
    countActiveText: string;
    countLeaderboardText: string;
}

export default class Footer extends Component<Props> {
    static contextType = AppContext;

    render(props: Props) {
        var placeActiveVisible: boolean = true;
        var countActiveVisible: boolean = true;
        var placeLeaderboardVisible: boolean = true;
        var countLeaderboardVisible: boolean = true;
        switch (this.context.buttonPhase) {
            case ButtonPhase.Idle: {
                placeActiveVisible = false;
                placeLeaderboardVisible = false;
                break;
            }
            case ButtonPhase.Push:
            case ButtonPhase.Hold: {
                placeLeaderboardVisible = false;
                countLeaderboardVisible = false;
                break;
            }
            case ButtonPhase.Release: {
                placeActiveVisible = false;
                countActiveVisible = false;
                break;
            }
        }
        return (
            <footer class={style.footer}>
                <div class={style.counters}>
                    {placeActiveVisible && <Counter value={this.context.placeActiveValue} label={props.placeActiveText} />}
                    {countActiveVisible && <Counter value={this.context.countActiveValue} label={props.countActiveText} />}
                    {placeLeaderboardVisible && <Counter value={this.context.placeLeaderboardValue} label={props.placeLeaderboardText} />}
                    {countLeaderboardVisible && <Counter value={this.context.countLeaderboardValue} label={props.countLeaderboardText} />}
                </div>
                <h2 class={style.label}>{this.context.footerText}</h2>
            </footer>
        );
    }
}