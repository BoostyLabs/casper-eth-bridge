import MetaMaskOnboarding from '@metamask/onboarding';
import { useMemo } from 'react';

import { Modal } from '@app/components/common/Modal';

import appConfig from '@/app/configs/appConfig.json';
import { LocalStorageKeys, useLocalStorage } from '@/app/hooks/useLocalStorage';
import { NotificationsPlugin as Notifications } from '@app/plugins/notifications';
import { CasperWallet } from '@/wallets/casperWallet';
import { MetaMaskWallet } from '@/wallets/metamaskWallet';
import { PhantomWallet } from '@/wallets/phantomWallet';
import { WalletsService } from '@/wallets/service';

import casper from '@static/images/casper.svg';
import metamask from '@static/images/metamask.svg';
import { Cross } from '@static/images/svg';
import phantom from '@static/images/phantom-logo.svg';

import './index.scss';

/** Describes wallet fields required for setting in LocalStorage. */
class LocalStorageWalletParameters {
    public constructor(
        public address: string,
        public addressType: LocalStorageKeys,
        public isConnectedWalletType: LocalStorageKeys,
        public signatureType: LocalStorageKeys,
        public signature: string,
    ) {}
};

export type WalletsModalProps = {
    isOpen: boolean;
    onClose: () => void;
    walletAddresses: { casper: string, metamask: string, phantom: string };
    setWalletAddresses: (walletAddresses: { casper: string, metamask: string, phantom: string }) => void;
};

export const WalletsModal: React.FC<WalletsModalProps> = ({ isOpen, onClose, setWalletAddresses, walletAddresses }) => {
    const { setLocalStorageItem, getLocalStorageItem } = useLocalStorage();
    const casperWallet: CasperWallet = new CasperWallet();
    const metaMaskWallet: MetaMaskWallet = new MetaMaskWallet();
    const phantomWallet: PhantomWallet = new PhantomWallet();
    const metaMaskService = useMemo(() => new WalletsService(metaMaskWallet), []);
    const phantomService = useMemo(() => new WalletsService(phantomWallet), []);
    const casperService = useMemo(() => new WalletsService(casperWallet), []);
    const metamaskOnboarding = useMemo(() => new MetaMaskOnboarding(), []);
    const isCasperConnected = getLocalStorageItem(LocalStorageKeys.isCasperConnected);
    const isMetamaskConnected = getLocalStorageItem(LocalStorageKeys.isMetamaskConnected);
    const isPhantomConnected = getLocalStorageItem(LocalStorageKeys.isPhantomConnected);

    const setWalletInfoToLocalStorage = (
        { signature, signatureType, isConnectedWalletType, addressType, address }: LocalStorageWalletParameters): void => {
        setLocalStorageItem(signatureType, signature);
        setLocalStorageItem(isConnectedWalletType, true);
        setLocalStorageItem(addressType, address);
        Notifications.walletSuccessfullyConnected();
        onClose();
    }

    const connectWithMetamask = async() => {
        if (!MetaMaskOnboarding.isMetaMaskInstalled()) {
            metamaskOnboarding.startOnboarding();

            return;
        }
        try {
            await metaMaskService.connect();
            const signature = await metaMaskService.sign(appConfig.strings.AUTHENTICATION_MESSAGE);
            const modifiedSignature = signature.slice(appConfig.numbers.TWO_NUMBER);
            const metamaskWalletAddress = await metaMaskService.address();

            setWalletInfoToLocalStorage(
                new LocalStorageWalletParameters(
                    metamaskWalletAddress,
                    LocalStorageKeys.metamaskAddress,
                    LocalStorageKeys.isMetamaskConnected,
                    LocalStorageKeys.metamaskSignature,
                    modifiedSignature,
                )
            );
            location.reload();
        } catch (error: any) {
            switch (error.message) {
                case appConfig.strings.USER_REJECTED_REQUEST:
                    Notifications.couldNotConnectViaMetaMask();
                    return;
                case appConfig.strings.metamaskErrors.USER_DENIED_MESSAGE_SIGNATURE:
                    Notifications.couldNotAuthenficateViaMetaMask();
                    return;
                default:
                    Notifications.couldNotGetMetamaskWalletAddres();
                    break;
            }
        }
    };

    const connectWithCasper = async() => {
        try {
            /** Checks if casper signer scripts injected. */
            if (!window?.casperlabsHelper) {
                window.open('https://chrome.google.com/webstore/detail/casper-signer/djhndpllfiibmcdbnmaaahkhchcoijce', '_blank');
                return;
            }
            const isConnected = await window?.casperlabsHelper?.isConnected();

            if (!isConnected) {
                await casperService.connect();
                return;
            }

            const publicKey = await casperService.address();
            const signature = await casperService.sign(appConfig.strings.AUTHENTICATION_MESSAGE);

            setWalletInfoToLocalStorage(
                new LocalStorageWalletParameters(
                    publicKey,
                    LocalStorageKeys.casperPublicKey,
                    LocalStorageKeys.isCasperConnected,
                    LocalStorageKeys.casperSignature,
                    signature,
                )
            );
            location.reload();
        } catch (error: any) {
            switch (error.message) {
                case appConfig.strings.CASPER_UNLOCK_ERROR_MESSAGE:
                    Notifications.unlockCasperSigner();
                    return;
                case appConfig.strings.CONNECT_CASPER_ERROR:
                    Notifications.connectCasperError();
                    return;
                case appConfig.strings.CANCEL_CASPER_ACTION:
                    return;
                default:
                    Notifications.couldNotGetCasperWalletAddres();
                    break;
            }
        }
    };

    const connectWithPhantom = async() => {
        try {
            // @ts-ignore
            if (!window.phantom?.solana.isPhantom) {
                window.open('https://phantom.app/', '_blank');
                return;
            }

            const phantomPublicKey = await phantomService.address();
            const signature = await phantomService.sign(appConfig.strings.AUTHENTICATION_MESSAGE);

            setWalletInfoToLocalStorage(
                new LocalStorageWalletParameters(
                    phantomPublicKey,
                    LocalStorageKeys.metamaskAddress,
                    LocalStorageKeys.isPhantomConnected,
                    LocalStorageKeys.metamaskSignature,
                    signature,
                )
            );
        } catch (error: any) {
            switch (error.message) {
                case appConfig.strings.USER_REJECTED_REQUEST:
                    Notifications.couldNotConnectViaPhantom();
                    return;
                // TODO: Will be added another error cases.
                default:
                    Notifications.couldNotGetPhantomWalletPublicKey();
                    break;
            }
        }
    }

    return (
        <Modal isOpen={isOpen} onClose={onClose}>
            <div className="wallets__wrapper">
                <div className="wallets__header">
                    <span className="wallets__header__title">Select Wallet</span>
                    <div className="wallets__header__close" onClick={onClose}>
                        <Cross />
                    </div>
                </div>
                <div className="wallets__body">
                    <button className="wallets__body__item" onClick={connectWithMetamask} disabled={isMetamaskConnected}>
                        <img className="wallets__body__item__logo" src={metamask} alt="metamask" />
                        <span className="wallets__body__item__name">Metamask</span>
                    </button>
                    <button className="wallets__body__item" onClick={connectWithCasper} disabled={isCasperConnected}>
                        <img className="wallets__body__item__logo" src={casper} alt="casper" />
                        <span className="wallets__body__item__name">Casper Wallet</span>
                    </button>
                    {/** TODO: Will be added later. */}
                    {/* <button className="wallets__body__item" onClick={connectWithPhantom} disabled={isPhantomConnected}>
                        <img className="wallets__body__item__logo" src={phantom} alt="phantom" />
                        <span className="wallets__body__item__name">Phantom Wallet</span>
                    </button> */}
                </div>
            </div>
        </Modal>
    )
}
