import App from './app';
import { Telegram } from "@twa-dev/types";
import { h } from 'preact';

declare global {
    interface Window {
        Telegram: Telegram;
    }
}

export default function Root() {
    // Get parameters from meta data of document
    const debug_el: Element = document.querySelector('[name="debug"]');
    const debug: boolean = JSON.parse(debug_el?.getAttribute('content') || "false");
    const client_id_el: Element = document.querySelector('[name="client_id"]');
    const client_id: string = client_id_el?.getAttribute('content') || "";
    const launch_params_el: Element = document.querySelector('[name="launch_params"]');
    const launch_params: string = launch_params_el?.getAttribute('content') || "";
    const api_host_el: Element = document.querySelector('[name="api_host"]');
    const api_host: string = api_host_el?.getAttribute('content') || window.location.host;

    return (
        <App
            apiHost={api_host}
            clientId={client_id}
            debug={debug}
            launchParams={launch_params}
            webApp={window.Telegram.WebApp}
        />
    );
}