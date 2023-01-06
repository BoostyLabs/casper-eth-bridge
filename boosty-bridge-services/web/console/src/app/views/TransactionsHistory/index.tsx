import { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { EmptyState } from '@app/components/common/EmptyState';
import { Paginator } from '@app/components/common/Paginator';
import { Search } from '@app/components/common/Search';
import { TransferItem } from '@app/components/transactionsHistory/TransferItem';

import appConfig from '@/app/configs/appConfig.json';
import { LocalStorageKeys, useLocalStorage } from '@/app/hooks/useLocalStorage';
import { NotificationsPlugin as Notifications } from '@/app/plugins/notifications';
import { cancelTransfer, getTransfersHistory } from '@/app/store/actions/transfers';
import { RootState } from '@/app/store';
import { Networks } from '@/networks';
import { Transfer, TransferPagination, TransferStatuses } from '@/transfers';

import './index.scss';

const TRANSACTIONS_FILTERS_CONFIG: string[] = [
    '#',
    'From',
    'To',
    'Destination',
    'Amount',
    'Time',
    'Date',
    'State',
    'Action',
];

const TransactionsHistory: React.FC = () => {
    const dispatch = useDispatch();
    const { getLocalStorageItem } = useLocalStorage();
    const { transfers, totalCount } = useSelector((state: RootState) => state.transfersReducer.history);
    const casperSignature: string = getLocalStorageItem(LocalStorageKeys.casperSignature);
    const metamaskSignature: string = getLocalStorageItem(LocalStorageKeys.metamaskSignature);
    const casperPublicKey: string = getLocalStorageItem(LocalStorageKeys.casperPublicKey);
    const isCasperConnected = getLocalStorageItem(LocalStorageKeys.isCasperConnected);
    const isMetamaskConnected = getLocalStorageItem(LocalStorageKeys.isMetamaskConnected);
    const areBothWalletsConnected: boolean = !!(isCasperConnected && isMetamaskConnected);
    const [activeNetwork, setActiveNetwork] = useState<Networks>(Networks.ETH);
    const [transferPagination, setTransferPagination] = useState<TransferPagination>(
        new TransferPagination(
            metamaskSignature,
            metamaskSignature,
            Networks.ETH,
            appConfig.numbers.ZERO_NUMBER,
            appConfig.numbers.FIVE_NUMBER
        )
    );
    const [searchedWalletAddress, setSearchedWalletAddress] = useState<string>('');
    /** Indicates public key depends on active network. */
    const pubKey: string = activeNetwork ? metamaskSignature : casperPublicKey;
    /** Indicates active signature depends on active network. */
    const activeSignature: string | null = activeNetwork ? metamaskSignature : casperSignature;

    const getSwitchButtonClassName: (network: Networks) => string = (network) => {
        const mainButtonClassName: string = `transactions-history__menu__switch__button${areBothWalletsConnected ? '' : '-not-allowed'}`;
        const activeButtonClassName: string = network === activeNetwork ? 'active' : '';

        return `${mainButtonClassName} ${activeButtonClassName}`;
    };

    /** Changes active button depend on network. */
    const changeActiveSwitchButton = (network: Networks) => {
        setActiveNetwork(network);
    };

    const changeOffset = (offset: number) => {
        setTransferPagination({ ...transferPagination, offset });
    };

    const encodeTransferStatus = (transferStatus: TransferStatuses): number => {
        const textEncoder = new TextEncoder();
        return textEncoder.encode(transferStatus.toString())[appConfig.numbers.ZERO_NUMBER];
    };

    const copyWalletAddress = (walletAddress: string) => {
        navigator.clipboard.writeText(walletAddress);
        Notifications.copied();
    };

    const changeSearchedWalletAddress = (e: React.ChangeEvent<HTMLInputElement>) => {
        setSearchedWalletAddress(e.target.value);
    };

    /** Canceles transfer and closes popup. */
    const cancelCurrentTransfer = async(id: number) => {
        try {
            const signature = activeNetwork ? metamaskSignature : casperSignature;
            await dispatch(cancelTransfer(id, signature, pubKey));
            await dispatch(getTransfersHistory(transferPagination));
            Notifications.transferSuccessfullyCanceled();
        } catch (error) {
            Notifications.couldNotCancelTransfer();
        }
    };

    useEffect(() => {
        if (!activeSignature) {
            return;
        };

        (async() => {
            try {
                await dispatch(getTransfersHistory(transferPagination));
            } catch (error) {
                Notifications.couldNotGetTransfersHistory()
            }
        })();
    }, [transferPagination]);

    useEffect(() => {
        /** Set's ETH network as default if both wallets connected. */
        if (areBothWalletsConnected) {
            setActiveNetwork(Networks.ETH);
            return;
        }

        if (isCasperConnected) {
            setActiveNetwork(Networks.CASPER);
        }
    }, []);

    /** Reset's transfer pagination if active network was changed. */
    useEffect(() => {
        if (!activeSignature) {
            return;
        };

        setTransferPagination(new TransferPagination(
            pubKey,
            activeSignature,
            activeNetwork,
            appConfig.numbers.ZERO_NUMBER,
            appConfig.numbers.FIVE_NUMBER
        ));
    }, [activeNetwork]);

    return <>
        <div className="transactions-history">
            <div className="transactions-history__menu">
                <div className="transactions-history__menu__switch">
                    <button
                        aria-label="Switch to ETH"
                        className={getSwitchButtonClassName(Networks.ETH)}
                        onClick={() => changeActiveSwitchButton(Networks.ETH)}
                        disabled={!areBothWalletsConnected}
                    >
                        ETH
                    </button>
                    <button
                        aria-label="Switch to Casper"
                        className={getSwitchButtonClassName(Networks.CASPER)}
                        onClick={() => changeActiveSwitchButton(Networks.CASPER)}
                        disabled={!areBothWalletsConnected}
                    >
                        CASPER
                    </button>
                </div>
                <Search value={searchedWalletAddress} changeValue={changeSearchedWalletAddress} />
            </div>
            <ul className="transactions-history__filters">
                {
                    TRANSACTIONS_FILTERS_CONFIG.map((transactionsFilter: string) =>
                        <li
                            className="transactions-history__filters__item"
                            key={transactionsFilter}
                        >
                            {transactionsFilter}
                        </li>
                    )
                }
            </ul>
            <ul className="transactions-history__list">
                {
                    transfers.length ?
                        transfers.map((transfer: Transfer, index: number) =>
                            <TransferItem
                                cancelTransfer={cancelCurrentTransfer}
                                copyWalletAddress={copyWalletAddress}
                                encodeTransferStatus={encodeTransferStatus}
                                transfer={transfer}
                                key={index}
                            />
                        ) : <EmptyState />
                }
            </ul>
        </div>
        {
            Boolean(totalCount) && <Paginator itemsCount={totalCount} changeOffset={changeOffset} />
        }
    </>;
};

export default TransactionsHistory;
