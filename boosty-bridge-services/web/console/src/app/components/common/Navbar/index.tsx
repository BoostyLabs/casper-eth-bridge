import { useEffect, useMemo, useState } from 'react';
import { Link } from 'react-router-dom';

import { WalletsModal } from '@app/components/SelectWalletsModal';

import appConfig from '@/app/configs/appConfig.json';
import { useTheme } from '@/app/hooks/useTheme';
import { RoutesConfig } from '@/app/routes';
import { LocalStorageKeys, useLocalStorage } from '@/app/hooks/useLocalStorage';
import { NotificationsPlugin as Notifications } from '@app/plugins/notifications';
import { MetaMaskWallet } from '@/wallets/metamaskWallet';
import { WalletsService } from '@/wallets/service';

import copy from '@static/images/common/copy.svg';
import helpDark from '@static/images/common/help.svg';
import helpGray from '@static/images/common/helpGray.svg';
import { Bridge, Explorer, History, Luna, Sun, Templates } from '@static/images/svg';

import './index.scss';

export class NavigationItem {
    constructor(
        public svg: () => JSX.Element,
        public label: string = '',
        public link: string = '',
    ) { };
};

const NAVIGATION_CONFIG: NavigationItem[] = [
    new NavigationItem(Bridge, 'Bridge', RoutesConfig.Swap.path),
    new NavigationItem(History, 'History', RoutesConfig.TransactionsHistory.path),
    new NavigationItem(Templates, 'Templates'),
    new NavigationItem(Explorer, 'Explorer'),
];

class WalletAddreses {
    constructor(
        public casper: string = '',
        public metamask: string = '',
        public phantom: string = '',
    ) {};
};

export const Navbar: React.FC = () => {
    const { isDarkModeOn, changeThemeMode } = useTheme();
    const metaMaskWallet: MetaMaskWallet = new MetaMaskWallet();
    const metaMaskService = useMemo(() => new WalletsService(metaMaskWallet), []);
    const { getLocalStorageItem } = useLocalStorage();
    const [walletAddresses, setWalletAddresses] = useState<WalletAddreses>(new WalletAddreses());
    const [isWalletsModalOpen, setIsWalletsModalOpen] = useState<boolean>(false);
    const darkModeMainLabelClassName: string = `navbar__additional__button-dark-mode__image-${isDarkModeOn ? 'black' : 'white'}`;
    const darkModeAdditionalLabelClassName: string = `navbar__additional__button-dark-mode__image-${isDarkModeOn ? 'white' : 'black'}`;
    const isMetamaskConnected = getLocalStorageItem(LocalStorageKeys.isMetamaskConnected);
    const isCasperConnected = getLocalStorageItem(LocalStorageKeys.isCasperConnected);
    const casperPublicKey = getLocalStorageItem(LocalStorageKeys.casperPublicKey);

    const getWalletAddressLabel = (address: string): string => {
        return `${address.slice(appConfig.numbers.ZERO_NUMBER, appConfig.numbers.FOUR_NUMBER)}...${address.slice(appConfig.numbers.MINUS_FOUR_NUMBER)}`
    };

    const copyWalletAddress = (address: string) => {
        navigator.clipboard.writeText(address);
        Notifications.copied();
    };

    const getWalletAddres = async() => {
        const updatedWalletAddress = new WalletAddreses();
        if (isMetamaskConnected) {
            try {
                await metaMaskService.connect();
                const metamaskWalletAddress: string = await metaMaskService.address();
                updatedWalletAddress.metamask = metamaskWalletAddress;
            } catch (error) {
                Notifications.couldNotGetMetamaskWalletAddres();
            }
        }

        if (isCasperConnected) {
            updatedWalletAddress.casper = casperPublicKey;
        }
        setWalletAddresses(updatedWalletAddress);
    };

    useEffect(() => {
        (async() => await getWalletAddres())();
    }, []);

    return <nav className="navbar">
        <div className="navbar__links-wrapper">
            <ul className="navbar__links">
                {
                    NAVIGATION_CONFIG.map((navigationItem: NavigationItem) =>
                        <li
                            key={navigationItem.label}
                            className={`navbar__links__item${navigationItem.link ? '' : '-not-allowed'}`}
                        >
                            {
                                navigationItem.link ?
                                    <Link
                                        className="navbar__links__item__label"
                                        to={navigationItem.link}
                                    >
                                        <navigationItem.svg />
                                        {navigationItem.label}
                                    </Link> : <span className="navbar__links__item__label">
                                        <navigationItem.svg />
                                        {navigationItem.label}
                                    </span>
                            }
                        </li>
                    )
                }
            </ul>
        </div>
        <div className="navbar__additional">
            <button aria-label="change plan" className="navbar__additional__button">
                Change Plan
            </button>
            <button aria-label="Help" className="navbar__additional__button">
                Help
                <img
                    alt="help"
                    src={isDarkModeOn ? helpGray : helpDark }
                    className="navbar__additional__button__help"
                />
            </button>
            <button
                aria-label="Copy"
                className="navbar__additional__button"
                onClick={isMetamaskConnected ? () => copyWalletAddress(walletAddresses.metamask) : () => setIsWalletsModalOpen(true)}
            >
                {isMetamaskConnected ? getWalletAddressLabel(walletAddresses.metamask) : 'Connect'}
                {
                    isMetamaskConnected && <img
                        alt="copy"
                        src={copy}
                        className="navbar__additional__button__copy"
                    />
                }
            </button>
            {/** TODO: This button added for MVP version. */}
            {(isMetamaskConnected || isCasperConnected) &&
                <button
                    aria-label="Copy"
                    className="navbar__additional__button"
                    onClick={isCasperConnected ? () => copyWalletAddress(walletAddresses.casper) : () => setIsWalletsModalOpen(true)}
                >
                    {isCasperConnected ? getWalletAddressLabel(walletAddresses.casper) : 'Connect another wallet'}
                    {
                        isCasperConnected && <img
                            alt="copy"
                            src={copy}
                            className="navbar__additional__button__copy"
                        />
                    }
                </button>
            }
            <button
                aria-label="Switch mode"
                onClick={changeThemeMode}
                className="navbar__additional__button-dark-mode"
            >
                <div className={darkModeMainLabelClassName}>
                    <Luna />
                </div>
                <div className={darkModeAdditionalLabelClassName}>
                    <Sun />
                </div>
            </button>
        </div>
        <WalletsModal
            isOpen={isWalletsModalOpen}
            onClose={() => setIsWalletsModalOpen(false)}
            setWalletAddresses={setWalletAddresses}
            walletAddresses={walletAddresses}
        />
    </nav>;
};
