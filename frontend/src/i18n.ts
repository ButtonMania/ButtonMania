import i18next from 'i18next';
import ui_en from './locales/en.json';
import ui_ru from './locales/ru.json';

i18next.init({
    debug: true,
    fallbackLng: 'en',
    defaultNS: 'ui',
    resources: {
        en: {
            ui: ui_en
        },
        ru: {
            ui: ui_ru
        }
    },
});

export default i18next;