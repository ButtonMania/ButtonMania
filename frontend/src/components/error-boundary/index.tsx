import { Component, ComponentChild, h } from 'preact';
import style from './style.css';

interface Props {
    children: ComponentChild;
}

interface State {
    hasError: boolean;
    errorMessage: string;
}

export default class ErrorBoundary extends Component<Props, State> {
    constructor(props: Props) {
        super(props);
        this.state = { hasError: false, errorMessage: '' };
    }

    static getDerivedStateFromError(error: Error): State {
        return { hasError: true, errorMessage: error.message || 'An unexpected error occurred' };
    }

    componentDidCatch(error: Error): void {
        console.error('ErrorBoundary caught:', error);
    }

    handleRetry = (): void => {
        this.setState({ hasError: false, errorMessage: '' });
    };

    render(props: Props, state: State): ComponentChild {
        if (state.hasError) {
            return (
                <div class={style.error_container} role="alert" aria-live="assertive">
                    <h2 class={style.error_title}>Something went wrong</h2>
                    <p class={style.error_message}>{state.errorMessage}</p>
                    <button
                        class={style.retry_button}
                        onClick={this.handleRetry}
                        aria-label="Retry loading the application"
                    >
                        Try Again
                    </button>
                </div>
            );
        }
        return props.children;
    }
}
