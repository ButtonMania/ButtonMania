import { ButtonPhase, ButtonType } from '../../protocol/enums';
import { Component, h } from 'preact';

import { AppContext } from '../context';
import style from './style.css';

interface Props {
}

interface State {
    loadHandler: any;
    fullyLoaded: boolean;
}

export default class Container extends Component<Props, State> {
    static contextType = AppContext;

    componentDidMount() {
        if (document.readyState === 'complete') {
            this.setState({ fullyLoaded: false }, () => this.loadHandler());
            return;
        }
        let loadHandler = this.loadHandler.bind(this);
        window.addEventListener('load', loadHandler);
        this.setState({
            loadHandler: loadHandler,
            fullyLoaded: false,
        });
    }

    componentWillUnmount() {
        if (this.state.loadHandler) {
            window.removeEventListener('load', this.state.loadHandler);
        }
    }

    loadHandler(_e?: Event): void {
        // Change state after little delay
        const fullyLoadedTimeout: number = 500;
        let fullyLoadedTimeoutId: any = window.setTimeout(() => {
            window.clearTimeout(fullyLoadedTimeoutId);
            this.setState({
                fullyLoaded: true,
            });
        }, fullyLoadedTimeout);
    }

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

    render(props: Props) {
        let fullyLoaded: boolean = this.state.fullyLoaded;
        let appClasses: string[] = [style.app];
        if (this.context.buttonPhase != ButtonPhase.Idle) {
            appClasses.push(this.buttonStyleClass(this.context.buttonType));
        }
        if (!fullyLoaded) {
            appClasses.push(style.loading);
        }
        return (
            <main class={appClasses.join(' ')}>
                <div class={style.outer}>
                    <div class={style.wrapper}>{fullyLoaded && this.props.children}</div>
                </div>
            </main>
        );
    }
}