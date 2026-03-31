import { Component, h } from 'preact';
import style from './style.css';

interface Props {}

interface State {
    isOnline: boolean;
    showBanner: boolean;
}

export default class ConnectionStatus extends Component<Props, State> {
    private onlineHandler: () => void;
    private offlineHandler: () => void;
    private hideTimeout: number | null = null;

    constructor(props: Props) {
        super(props);
        this.state = {
            isOnline: typeof navigator !== 'undefined' ? navigator.onLine : true,
            showBanner: false,
        };
        this.onlineHandler = this.handleOnline.bind(this);
        this.offlineHandler = this.handleOffline.bind(this);
    }

    componentDidMount(): void {
        window.addEventListener('online', this.onlineHandler);
        window.addEventListener('offline', this.offlineHandler);
    }

    componentWillUnmount(): void {
        window.removeEventListener('online', this.onlineHandler);
        window.removeEventListener('offline', this.offlineHandler);
        if (this.hideTimeout) {
            window.clearTimeout(this.hideTimeout);
        }
    }

    handleOnline(): void {
        this.setState({ isOnline: true, showBanner: true });
        this.hideTimeout = window.setTimeout(() => {
            this.setState({ showBanner: false });
        }, 3000);
    }

    handleOffline(): void {
        this.setState({ isOnline: false, showBanner: true });
        if (this.hideTimeout) {
            window.clearTimeout(this.hideTimeout);
            this.hideTimeout = null;
        }
    }

    render(_: Props, state: State) {
        if (!state.showBanner) {
            return null;
        }

        const classes = [
            style.banner,
            state.isOnline ? style.online : style.offline,
        ];

        return (
            <div
                class={classes.join(' ')}
                role="status"
                aria-live="polite"
                aria-label={state.isOnline ? 'Connection restored' : 'No internet connection'}
            >
                {state.isOnline ? 'Back online' : 'No connection'}
            </div>
        );
    }
}
