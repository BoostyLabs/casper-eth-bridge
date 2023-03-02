import { useEffect, useMemo, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { EmptyState } from '@app/components/common/EmptyState';
import { Paginator } from '@app/components/common/Paginator';
import { Search } from '@app/components/common/Search';
import { TransferItem } from '@app/components/transactionsHistory/TransferItem';

import appConfig from '@/app/configs/appConfig.json';
import { LocalStorageKeys, useLocalStorage } from '@/app/hooks/useLocalStorage';
import { NotificationsPlugin as Notifications } from '@/app/plugins/notifications';
import { getConnectedNetworks } from '@app/store/actions/networks';
import { getTransfersHistory, setHistory } from '@/app/store/actions/transfers';
import { RootState } from '@/app/store';
import { Network, NetworkNames, NetworkTypes } from '@/networks';
import { CancelSignatureRequest, Transfer, TransferPagination, TransferStatuses, TransfersHistory } from '@/transfers';
import { CasperWallet } from '@/wallets/casperWallet';
import { MetaMaskWallet } from '@/wallets/metamaskWallet';
import { WalletsService } from '@/wallets/service';

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
    const networks = useSelector((state: RootState) => state.networksReducer.networks);
    const casperSignature: string = getLocalStorageItem(LocalStorageKeys.casperSignature);
    const metamaskSignature: string = getLocalStorageItem(LocalStorageKeys.metamaskSignature);
    const casperPublicKey: string = getLocalStorageItem(LocalStorageKeys.casperPublicKey);
    const isCasperConnected = getLocalStorageItem(LocalStorageKeys.isCasperConnected);
    const isMetamaskConnected = getLocalStorageItem(LocalStorageKeys.isMetamaskConnected);
    const areBothWalletsConnected: boolean = !!(isCasperConnected && isMetamaskConnected);
    const [activeNetwork, setActiveNetwork] = useState<Network>(new Network());
    const [transferPagination, setTransferPagination] = useState<TransferPagination>(
        new TransferPagination(
            metamaskSignature,
            metamaskSignature,
            activeNetwork.id,
            appConfig.numbers.ZERO_NUMBER,
            appConfig.numbers.FIVE_NUMBER
        )
    );
    const [searchedWalletAddress, setSearchedWalletAddress] = useState<string>('');
    const metaMaskWallet: MetaMaskWallet = new MetaMaskWallet();
    const metaMaskService = useMemo(() => new WalletsService(metaMaskWallet), []);
    const casperWallet: CasperWallet = new CasperWallet();
    const casperServise = useMemo(() => new WalletsService(casperWallet), []);

    /** Request network by type, i.e CASPER, EVM. */
    const getNetworkByType = (searchedNetwork: NetworkTypes) => {
        const network = networks && networks.find((network: Network) => network.type === searchedNetwork)
        if (!network) {
            return new Network();
        }
        return network;
    };

    /** Indicates public key depends on active network. */
    const pubKey: string = getNetworkByType(NetworkTypes.EVM).type === activeNetwork.type ? metamaskSignature : casperPublicKey;
    /** Indicates active signature depends on active network. */
    const activeSignature: string | null = getNetworkByType(NetworkTypes.EVM).type === activeNetwork.type ? metamaskSignature : casperSignature;

    const getSwitchButtonClassName: (networkName: NetworkNames) => string = (networkName) => {
        const mainButtonClassName: string = 'transactions-history__menu__switch__button';
        const activeButtonClassName: string = networkName === activeNetwork.name ? 'active' : '';

        return `${mainButtonClassName} ${activeButtonClassName}`;
    };

    /** Changes active network. */
    const changeActiveNetwork = (network: Network) => {
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
    const cancelTransfer = async(transferId: number, address: string) => {
        if (activeNetwork.type === NetworkTypes.CASPER) {
            // TODO: implement.
            return;
        }
        try {
            const signature = metamaskSignature;
            const cancelSignatureRequest = new CancelSignatureRequest(
                transferId,
                signature,
                activeNetwork.id,
                address,
            );
            await metaMaskService.cancelTransaction(cancelSignatureRequest);
            await dispatch(getTransfersHistory(transferPagination));
            Notifications.transferSuccessfullyCanceled();
        } catch (error) {
            Notifications.couldNotCancelTransfer();
        }
    };

    useEffect(() => {
        if (!activeSignature) {
            dispatch(setHistory(new TransfersHistory()));
            return;
        };

        (async() => {
            try {
                transferPagination.networkId && await dispatch(getTransfersHistory(transferPagination));
            } catch (error) {
                Notifications.couldNotGetTransfersHistory();
            }
        })();
    }, [transferPagination]);

    useEffect(() => {
        /** Set's ETH network as default if both wallets connected. */
        if (areBothWalletsConnected) {
            setActiveNetwork(getNetworkByType(NetworkTypes.EVM));
            return;
        }

        if (isMetamaskConnected) {
            setActiveNetwork(getNetworkByType(NetworkTypes.EVM));
        }

        if (isCasperConnected) {
            setActiveNetwork(getNetworkByType(NetworkTypes.CASPER));
        }
    }, [networks]);

    useEffect(() => {
        (async() => {
            try {
                await dispatch(getConnectedNetworks());
            } catch (e: any) {
                Notifications.couldNotGetConnectedNetworks();
            }
        })();
    }, []);

    /** Reset's transfer pagination if active network was changed. */
    useEffect(() => {
        if (!activeSignature) {
            dispatch(setHistory(new TransfersHistory()));
            return;
        };

        setTransferPagination(new TransferPagination(
            pubKey,
            activeSignature,
            activeNetwork.id,
            appConfig.numbers.ZERO_NUMBER,
            appConfig.numbers.FIVE_NUMBER
        ));
    }, [activeNetwork.id]);

    return <>
        <div className="transactions-history">
            <div className="transactions-history__menu">
                <div className="transactions-history__menu__switch">
                    {
                        networks.map(network =>
                            <button
                                key={network.id}
                                aria-label={`Switch to ${network.name}`}
                                className={getSwitchButtonClassName(network.name)}
                                onClick={() => changeActiveNetwork(network)}
                            >
                                {network.name}
                            </button>
                        )
                    }
                </div>
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
                                cancelTransfer={cancelTransfer}
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
