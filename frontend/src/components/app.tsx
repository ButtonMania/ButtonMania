import { AppContext, AppContextProps } from './context';
import { ButtonPhase, ButtonType, UserLocale } from '../protocol/enums';
import { Component, ComponentChild, h } from 'preact';
import { ConnectEvent, DisconnectEvent } from '../protocol/internal';
import { ErrorResponse, RecordResponse, RoomStats, UpdateResponse } from '../protocol/respons';
import { Platforms, WebApp, WebAppInitData } from "@twa-dev/types";
import { enumIndex, randomEnum } from '../utils';

import Button from './button';
import { Config } from '../config';
import Container from './container';
import Footer from './footer';
import { GameplayMessageRequest } from '../protocol/requests';
import Header from './header';
import { MessageBus } from '../services/bus';
import { NetworkClient } from '../services/client';
import i18next from '../i18n';

interface Props {
	apiHost: string;
	clientId: string;
	debug: boolean;
	launchParams: string;
	webApp: WebApp;
}

interface State extends AppContextProps {
	lastSendUpdateTimestamp: number;
}

export default class App extends Component<Props, State> {
	private readonly config: Config;
	private readonly client: NetworkClient;
	private readonly sendUpdateFrequency: number = 5600;

	constructor(public props: Props) {
		super(props);

		this.config = new Config(props.apiHost, props.clientId, props.debug);
		this.client = new NetworkClient(this.config);

		MessageBus.default(RoomStats).subscribe(this.onRecieveRoomStats.bind(this));
		MessageBus.default(UpdateResponse).subscribe(this.onRecieveUpdate.bind(this));
		MessageBus.default(RecordResponse).subscribe(this.onRecieveRecord.bind(this));
		MessageBus.default(ErrorResponse).subscribe(this.onRecieveError.bind(this));
		MessageBus.default(ConnectEvent).subscribe(this.onStartGameplay.bind(this));
		MessageBus.default(DisconnectEvent).subscribe(this.onStopGameplay.bind(this));
	}

	componentWillMount(): void {
		if (typeof window === 'undefined') {
			return;
		}
		let tgData: WebAppInitData = this.props.webApp.initDataUnsafe;
		var locale: UserLocale = UserLocale.EN;
		var userId: number = this.props.debug ? 0 : NaN;
		var isPremium: boolean = false;
		// Fetch data from telegram mini app api
		if (tgData != null && tgData.user != null) {
			locale = this.getUserLocal(tgData.user.language_code);
			userId = tgData.user.id;
			isPremium = tgData.user.is_premium;
		}
		// Detect new year special mode
		const now: Date = new Date();
		const isNewYear: boolean = (
			(now.getUTCMonth() >= 11 && now.getUTCDate() == 31) ||
			(now.getUTCMonth() == 0 && now.getUTCDate() <= 10)
		);
		// Set initial button type and state
		var maxType: number = enumIndex(ButtonType, isPremium ? ButtonType.Prestige : ButtonType.Fortune);
		var minType: number = enumIndex(ButtonType, isNewYear ? ButtonType.NewYear : ButtonType.Peace);
		let buttonType: ButtonType = isNewYear ? ButtonType.NewYear : randomEnum(ButtonType, minType, maxType);
		let buttonPhase: ButtonPhase = ButtonPhase.Idle;
		// Fetch initial room stats, update locale and adjust font size
		this.client.fetchRoomStats(buttonType);
		this.adjustBaseFontSize(this.props.webApp.platform);
		i18next.changeLanguage(locale);
		// Set initial state
		this.setState({
			initData: this.props.webApp.initData || this.props.launchParams,
			telegramUserID: userId,
			isPremium: isPremium,
			isNewYear: isNewYear,
			locale: locale,
			buttonPhase: buttonPhase,
			buttonType: buttonType,
			messageText: i18next.t(`${buttonType}DefaultSubtitle`),
			currentRecordText: i18next.t('currentRecordHeaderText'),
			currentRecordValue: 0,
			placeActiveValue: 0,
			placeLeaderboardValue: 0,
			countActiveValue: 0,
			countLeaderboardValue: 0,
			holdDuration: 0,
			worldRecord: false,
		}, this.onChangeButtonState);
	}

	componentDidMount(): void {
		// We are ready to launch app
		this.props.webApp.ready();
		this.props.webApp.expand();
	}

	adjustBaseFontSize(platform: Platforms): void {
		var fontSize = 16;
		switch (platform as string) {
			case "android":
			case "android_x":
			case "ios":
			case "unknown": {
				break;
			}
			case "web": {
				fontSize = 14;
				break;
			}
			case "macos":
			case "tdesktop":
			case "weba":
			case "webk":
			case "unigram": {
				fontSize = 10;
				break;
			}
		}
		document.documentElement.style.fontSize = `${fontSize}px`;
	}

	getUserLocal(language_code: string): UserLocale {
		if (language_code == 'ru') {
			return UserLocale.RU
		}
		return UserLocale.EN
	}

	onChangeButtonState(): void {
		var footerText: string = null;
		var buttonText: string = null;
		var messageText: string = null;
		let buttonType: ButtonType = this.state.buttonType;
		switch (this.state.buttonPhase) {
			case ButtonPhase.Idle: {
				footerText = i18next.t('idlePhaseFooterText');
				messageText = i18next.t(`${buttonType}DefaultSubtitle`);
				break
			}
			case ButtonPhase.Push:
			case ButtonPhase.Hold: {
				footerText = i18next.t('holdPhaseFooterText');
				buttonText = i18next.t('holdPhaseButtonText');
				break;
			}
			case ButtonPhase.Release: {
				footerText = i18next.t('releasePhaseFooterText');
				buttonText = i18next.t(`${buttonType}ButtonText`);
				break;
			}
		}
		this.setState({
			buttonAnimation: `/assets/animations/${buttonType}.lottie`,
			headerText: i18next.t(buttonType),
			footerText: footerText,
			buttonText: buttonText,
			messageText: messageText || this.state.messageText
		});
	}

	onChangeButtonPhase(buttonPhase: ButtonPhase, holdDuration: number): void {
		var lastSendUpdateTimestamp: number = this.state.lastSendUpdateTimestamp;
		var countActive: number = this.state.countActiveValue;
		var placeActive: number = this.state.placeActiveValue;
		var pushTimestamp: number = this.state.pushTimestamp;

		switch (buttonPhase) {
			case ButtonPhase.Idle: {
				this.client.stopGameplay();
				break
			}
			case ButtonPhase.Push: {
				countActive = 1;
				placeActive = 1;
				pushTimestamp = Math.floor(Date.now() / 1000) | 0;
				this.client.startGameplay(
					this.state.initData,
					this.state.buttonType,
				);
				break;
			}
			case ButtonPhase.Hold: {
				const now: number = performance.now();
				if (this.sendUpdateFrequency > (now - lastSendUpdateTimestamp)) {
					break;
				}
				lastSendUpdateTimestamp = now;
			}
			case ButtonPhase.Release: {
				MessageBus.raise(new GameplayMessageRequest(buttonPhase)).catch(console.log);
				break;
			}
		}

		this.setState({
			lastSendUpdateTimestamp: lastSendUpdateTimestamp,
			buttonPhase: buttonPhase,
			holdDuration: holdDuration,
			pushTimestamp: pushTimestamp,
			countActiveValue: countActive,
			placeActiveValue: placeActive,
			worldRecord: false,
		}, this.onChangeButtonState);
	}

	onChangeButtonType(buttonType: ButtonType): void {
		this.setState({
			buttonType: buttonType,
			buttonPhase: ButtonPhase.Idle,
			worldRecord: false,
			holdDuration: 0,
			pushTimestamp: 0,
		}, this.onChangeButtonState);
		this.client.fetchRoomStats(buttonType);
	}

	onRecieveRoomStats(stats: RoomStats): void {
		this.setState({
			countActiveValue: stats.countActive,
			countLeaderboardValue: stats.countLeaderboard,
			currentRecordValue: stats.bestDuration,
		});
		this.props.webApp.HapticFeedback.impactOccurred('soft')
	}

	onRecieveUpdate(respons: UpdateResponse): void {
		var holdDuration: number = this.state.holdDuration;
		if ((holdDuration - respons.context.duration) > 0) {
			holdDuration = respons.context.duration;
		}
		this.setState({
			holdDuration: holdDuration,
			pushTimestamp: respons.context.timestamp,
			messageText: respons.message || this.state.messageText,
			placeActiveValue: respons.placeActive,
			countActiveValue: respons.countActive,
		});
		if (respons.message != null) {
			this.props.webApp.HapticFeedback.notificationOccurred('success')
		}
	}

	onRecieveRecord(respons: RecordResponse): void {
		let buttonType: ButtonType = this.state.buttonType;
		this.setState({
			pushTimestamp: respons.record.timestamp,
			holdDuration: respons.record.duration,
			placeLeaderboardValue: respons.placeLeaderboard,
			countLeaderboardValue: respons.countLeaderboard,
			worldRecord: respons.worldRecord,
			messageText: i18next.t('recordHeaderText'),
			buttonText: i18next.t(respons.worldRecord ? 'worldRecordButtonText' : `${buttonType}ButtonText`),
		});
		if (respons.worldRecord) {
			this.props.webApp.HapticFeedback.impactOccurred('heavy')
		} else {
			this.props.webApp.HapticFeedback.notificationOccurred('success')
		}
	}

	onRecieveError(respons: ErrorResponse): void {
		this.props.webApp.showPopup({ message: `Oops! An error has occurred: ${respons.error.message}` })
		this.props.webApp.HapticFeedback.notificationOccurred('error')
	}

	onStartGameplay(): void {
		this.setState({
			messageText: i18next.t(`${this.state.buttonType}Message`),
			worldRecord: false,
		});
	}

	onStopGameplay(): void {
		this.props.webApp.HapticFeedback.impactOccurred('rigid')
	}

	render(props: Props, state: State): ComponentChild {
		return (
			<AppContext.Provider value={state}>
				<Container>
					<Header />
					<Button
						buttonHandler={this.onChangeButtonPhase.bind(this)}
						arrowsHandler={this.onChangeButtonType.bind(this)}
					/>
					<Footer
						placeActiveText={i18next.t('placeActiveText')}
						placeLeaderboardText={i18next.t('placeLeaderboardText')}
						countActiveText={i18next.t('countActiveText')}
						countLeaderboardText={i18next.t('countLeaderboardText')}
					/>
				</Container>
			</AppContext.Provider >
		)
	}
}
