import { ButtonPhase, ButtonType, UserLocale } from '../protocol/enums';

import { createContext } from 'preact';

export interface BaseContextProps {
    locale: UserLocale;
    initData: string;
    telegramUserID: number;
    isPremium: boolean;
    buttonType: ButtonType;
    buttonPhase: ButtonPhase;
}

export interface HeaderContextProps extends BaseContextProps {
    headerText: string;
    messageText: string;
}

export interface ButtonContextProps extends BaseContextProps {
    pushTimestamp: number;
    holdDuration: number;
    worldRecord: boolean;
    buttonText: string;
    buttonAnimation: string;
}

export interface FooterContextProps extends BaseContextProps {
    placeActiveValue: number;
    placeLeaderboardValue: number;
    countActiveValue: number;
    countLeaderboardValue: number;
    footerText: string;
}

export interface AppContextProps extends
    HeaderContextProps,
    ButtonContextProps,
    FooterContextProps {
}

export const AppContext = createContext({} as AppContextProps);