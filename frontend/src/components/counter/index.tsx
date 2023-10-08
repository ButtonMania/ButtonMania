import { Component, h } from 'preact';

import style from './style.css';

interface Props {
    value: number;
    label: string;
}

export default class Counter extends Component<Props> {
    render(props: Props) {
        return (
            <div class={style.counter}>
                <div class={style.value}>{props.value}</div>
                <div class={style.label}>{props.label}</div>
            </div>
        );
    }
}