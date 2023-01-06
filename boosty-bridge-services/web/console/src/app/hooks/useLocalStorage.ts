/** Defines used local storage keys. */
export enum LocalStorageKeys {
    metamaskSignature = 'METAMASK_SIGNATURE',
    casperSignature = 'CASPER_SIGNATURE',
    casperPublicKey = 'CASPER_PUBLIC_KEY',
    phantomPublicKey = 'PHANTOM_PUBLIC_KEY',
    metamaskAddress = 'METAMASK_ADDRESS',
    phantomSignature = 'PHANTOM_SIGNATURE',
    themeMode = 'THEME_MODE',
    isMetamaskConnected = 'IS_METAMASK_CONNECTED',
    isCasperConnected = 'IS_CASPER_CONNECTED',
    isPhantomConnected = 'IS_PHANTOM_CONNECTED'
};

/** Hook gets/sets/deletes local storage value. */
export const useLocalStorage = () => {
    /* Set value to localStorage */
    const setLocalStorageItem = (item: string, value: string | boolean) =>
        window.localStorage && window.localStorage.setItem(item, JSON.stringify(value));

    /* Get value from localStorage */
    const getLocalStorageItem = (item: string) => {
        const storageItem: string | null = window.localStorage && window.localStorage.getItem(item);

        return storageItem && JSON.parse(storageItem);
    };

    /* Remove item from localStorage */
    const removeLocalStorageItem = (item: string) =>
        window.localStorage && window.localStorage.removeItem(item);

    return { setLocalStorageItem, getLocalStorageItem, removeLocalStorageItem };
};
