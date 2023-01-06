import { useEffect, useState } from 'react';

import { LocalStorageKeys, useLocalStorage } from './useLocalStorage';

enum ThemeModes {
    dark = 'dark',
    ligh = 'light',
};

/** Hook to change themes mode. */
export const useTheme = () => {
    const { getLocalStorageItem, setLocalStorageItem } = useLocalStorage();
    const [theme, setTheme] = useState<ThemeModes>(getLocalStorageItem(LocalStorageKeys.themeMode));
    const [isDarkModeOn, setIsDarkModeOn] = useState<boolean>(theme === ThemeModes.dark);

    const changeThemeMode = () => {
        setIsDarkModeOn(!isDarkModeOn);
    };

    useEffect(() => {
        setLocalStorageItem(LocalStorageKeys.themeMode, isDarkModeOn ? ThemeModes.dark : ThemeModes.ligh);
        document.body.classList.add(isDarkModeOn ? ThemeModes.dark : ThemeModes.ligh);
        document.body.classList.remove(isDarkModeOn ? ThemeModes.ligh : ThemeModes.dark);

        setTheme(getLocalStorageItem(LocalStorageKeys.themeMode));
    }, [isDarkModeOn]);

    return { isDarkModeOn, changeThemeMode };
};
