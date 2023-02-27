import { useMemo, useState } from 'react';
import { useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import LinearProgress from '@mui/material/LinearProgress';

import { Modal } from '@app/components/common/Modal';

import appConfig from '@/app/configs/appConfig.json';
import { LocalStorageKeys, useLocalStorage } from '@app/hooks/useLocalStorage';
import { NotificationsPlugin as Notifications } from '@app/plugins/notifications';
import { RoutesConfig } from '@/app/routes';
import { RootState } from '@/app/store';
import { NetworkNames } from '@/networks';
import { CasperWallet } from '@/wallets/casperWallet';
import { MetaMaskWallet } from '@/wallets/metamaskWallet';
import { WalletsService } from '@/wallets/service';

import arrow from '@static/images/arrow.svg';
import favorite from '@static/images/favorite.svg';

import './index.scss';

type TransactionModalProps = {
    amount: string;
    asset: string;
    destination: string;
    fromTokenImg: string;
    isOpen: boolean;
    onClose: () => void;
    toTokenImg: string;
    activeNetwork: NetworkNames;
};

/** Describes each transaction item entity for list rendering. */
class TransactionItem {
    public constructor(
        public label: string = '',
        public value: number | string = '',
    ) {}
}

export const TransactionModal: React.FC<TransactionModalProps> = ({ amount, destination, isOpen, onClose, fromTokenImg, toTokenImg, activeNetwork }) => {
    const navigation = useNavigate();
    const { fee, feePercentage } = useSelector((state: RootState) => state.transfersReducer.transferEstimate);
    const { getLocalStorageItem } = useLocalStorage();
    const SENDER_NETWROK_ID: number = Number(getLocalStorageItem(LocalStorageKeys.senderNetworkId));
    const RECIPIENT_NETWORK_ID: number = Number(getLocalStorageItem(LocalStorageKeys.recipientNetworkId));
    const casperWallet: CasperWallet = new CasperWallet()
    const metaMaskWallet: MetaMaskWallet = new MetaMaskWallet();
    const metaMaskService = useMemo(() => new WalletsService(metaMaskWallet), [SENDER_NETWROK_ID, RECIPIENT_NETWORK_ID]);
    const casperService = useMemo(() => new WalletsService(casperWallet), [SENDER_NETWROK_ID, RECIPIENT_NETWORK_ID]);
    const [transactionProgress, setTransactionProgress] = useState<number>(appConfig.numbers.ZERO_NUMBER);
    const [isTransactionProgressStarted, setIsTransactionProgressStarted] = useState<boolean>(false);
    const [isTransactionFavorite, setIsTransactionFavorite] = useState<boolean>(false);
    const isCasperActiveNetwork: boolean = activeNetwork === NetworkNames.CASPER_TEST;
    const progressBarTextClassName = transactionProgress > appConfig.numbers.FIFTY_NUMBER ? '-white' : '';
    const progressBarClassName = transactionProgress === appConfig.numbers.ONE_HUNDRED_NUMBER ? '-finished' : '';
    const favoriteBlockClassName = isTransactionFavorite ? '-active' : '';
    const destinationLabel: string = `${destination.slice(appConfig.numbers.ZERO_NUMBER, appConfig.numbers.FOUR_NUMBER)}...${destination.slice(appConfig.numbers.MINUS_FOUR_NUMBER)}`;
    const [isConfirmButtonDisabled, setIsConfirmButtonDisabled] = useState<boolean>(false);

    const AMOUNT: number = Number(amount);
    const FEE: number = Number(fee);
    const TOTAL_PRICE: string = (AMOUNT + FEE).toFixed(appConfig.numbers.FOUR_NUMBER);

    /** Transaction values for rendering. */
    const transactionValues: TransactionItem[] = [
        new TransactionItem('Amount', amount),
        new TransactionItem('Destination', destinationLabel),
        new TransactionItem('Commission', `${feePercentage}%`),
        new TransactionItem('Total Price', TOTAL_PRICE),
    ];

    /** Approves and sends transaction. */
    const sendMetamaskTransaction = async() => {
        setIsConfirmButtonDisabled(true);
        try {
            await metaMaskService.sendTransaction(destination, amount);
            Notifications.success();
            onClose();
            navigation(RoutesConfig.TransactionsHistory.path);
            location.reload();
        } catch (error: any) {
            if (error.message === appConfig.strings.metamaskErrors.USER_DENIED_TRANSACTION_SIGNATURE) {
                setIsConfirmButtonDisabled(false);
                Notifications.transactionViaMetaMaskWasCanceled();
                return;
            }
            Notifications.couldNotSendTransactionViaMetaMaskWallet();
        };
    };

    /** Sign and send casper transaction. */
    const sendCasperTransaction = async() => {
        setIsConfirmButtonDisabled(true);
        setTransactionProgress(appConfig.numbers.ZERO_NUMBER);
        /** Check if the site is connected to Casper's extension. */
        const isConnected = await window?.casperlabsHelper?.isConnected();

        if (!isConnected) {
            /** Call the Casper extension for the site connection. */
            await casperService.connect();
        }

        try {
            await casperService.sendTransaction(amount, destination);
            // TODO: Will be added some timeout or websocket implementation to progress bar.
            setTransactionProgress(appConfig.numbers.ONE_HUNDRED_NUMBER);
            onClose();
            Notifications.transactionSuccess();
            navigation(RoutesConfig.TransactionsHistory.path);
            location.reload();
        } catch (error: any) {
            setIsConfirmButtonDisabled(false);
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
                    Notifications.couldNotSendTransactionViaCasperWallet();
                    break;
            }
        }
    }

    return (
        <Modal isOpen={isOpen} onClose={onClose}>
            <div className="transaction__wrapper">
                <div className="transaction__header">
                    <div className="transaction__header__token-from">
                        <img src={fromTokenImg} alt="token" />
                    </div>
                    <img src={arrow} alt="arrow" />
                    <div className="transaction__header__token-to">
                        <img src={toTokenImg} alt="token" />
                    </div>
                </div>
                <div className="transaction__body">
                    <div className="transaction__body__item">
                        <img src={fromTokenImg} alt="asset" />
                        <span className="transaction__body__label__asset">Asset</span>
                        <span className="transaction__body__value">ETH</span>
                    </div>
                    {
                        transactionValues.map((item: TransactionItem, index: number) =>
                            <div className="transaction__body__item" key={index}>
                                <span className="transaction__body__label">{item.label}</span>
                                <span className="transaction__body__value">{item.value}</span>
                            </div>
                        )
                    }
                </div>
                <div className="transaction__favorites">
                    <img className="transaction__favorites__icon" src={favorite} alt="favorite icon" />
                    <span className="transaction__favorites__label">Save to favorites</span>
                </div>
                {
                    transactionProgress === appConfig.numbers.ONE_HUNDRED_NUMBER &&
                        <div className="transaction__favorite" onClick={() => setIsTransactionFavorite(!isTransactionFavorite)}>
                            <div className="transaction__favorite__outer-checkbox">
                                <div className={`transaction__favorite__inner-checkbox${favoriteBlockClassName}`}></div>
                            </div>
                            <span className="transaction__favorite__label">Save to favorites</span>
                        </div>
                }
                {isTransactionProgressStarted
                    ?
                    <div className="transaction__progress-bar">
                        <span className={`transaction__progress-bar__text${progressBarTextClassName}`}>Loading...</span>
                        <LinearProgress variant="determinate" value={transactionProgress} className={`progress-bar${progressBarClassName}`} />
                        {/** TODO: Will be added Ok button and close transaction Modal */}
                    </div>
                    :
                    <div className="transaction__buttons">
                        <button aria-label="Cancel transaction" className="transaction__cancel-btn" onClick={onClose}>Cancel</button>
                        {/** TODO: Will be added wallet check and call different method */}
                        <button
                            aria-label="Confirm transaction"
                            className="transaction__confirm-btn"
                            disabled={isConfirmButtonDisabled}
                            onClick={isCasperActiveNetwork ? sendCasperTransaction : sendMetamaskTransaction}
                        >
                            Confirm
                        </button>
                    </div>
                }
            </div>
        </Modal>
    );
};
