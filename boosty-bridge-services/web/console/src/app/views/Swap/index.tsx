import { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { TransactionModal } from '@app/components/TransactionModal';
import { TokensModal } from '@app/components/TokensModal';
import { WalletsModal } from '@app/components/SelectWalletsModal';

import appConfig from '@/app/configs/appConfig.json';
import { LocalStorageKeys, useLocalStorage } from '@/app/hooks/useLocalStorage';
import { WalletAddressValidator } from '@/app/internal/walletAddressValidator';
import { NotificationsPlugin as Notifications } from '@/app/plugins/notifications';
import { RootState } from '@/app/store';
import { getConnectedNetworks } from '@/app/store/actions/networks';
import { estimateTransfer } from '@/app/store/actions/transfers';
import { Network, NetworkTypes } from '@/networks';
import { TransferEstimateRequest } from '@/transfers';

import download from '@app/static/images/download.svg';
import eth from '@app/static/images/eth.svg';
import swap from '@app/static/images/swap.svg';
import { Info } from '@app/static/images/svg/index';

import './index.scss';

enum SwapItems {
    'Token',
    'NFT'
}

const Swap: React.FC = () => {
    const dispatch = useDispatch();
    const { networks, activeSupportedToken } = useSelector((state: RootState) => state.networksReducer);
    const [isTransactionModalOpen, setIsTransactionModalOpen] = useState<boolean>(false);
    const [isTokensModalOpen, setIsTokensModalOpen] = useState<boolean>(false);
    const [selectedSwapItem, setSelectedSwapItem] = useState<SwapItems>(SwapItems.Token);
    const [isWalletsModalOpen, setIsWalletsModalOpen] = useState<boolean>(false);
    const [isSelectedCheckbox, setIsSelectedCheckbox] = useState<boolean>(false);
    const [isTokenReversed, setIsTokenReversed] = useState<boolean>(false);
    const [tokenAmount, setTokenAmount] = useState<string>('');
    const [destinationAddress, setDestinationAddress] = useState<string>('');
    const [firstTokenBlock, setFirstTokenBlock] = useState<Network>(new Network());
    const [secondTokenBlock, setSecondTokenBlock] = useState<Network>(new Network());
    const swapTokenClassName = selectedSwapItem === SwapItems.Token ? 'token-active' : 'token-inactive';
    const swapNftClassName = selectedSwapItem === SwapItems.NFT ? 'nft-active' : 'nft-inactive';
    const { getLocalStorageItem } = useLocalStorage();
    const [walletAddresses, setWalletAddresses] = useState<{ casper: string, metamask: string, phantom: string }>({ casper: '', metamask: '', phantom: '' });
    const isCasperConnected = getLocalStorageItem(LocalStorageKeys.isCasperConnected);
    const isMetamaskConnected = getLocalStorageItem(LocalStorageKeys.isMetamaskConnected);

    useEffect(() => {
        (async() => {
            try {
                await dispatch(getConnectedNetworks());
            } catch (e: any) {
                Notifications.couldNotGetConnectedNetworks();
            }
        })();
    }, []);

    useEffect(() => {
        networks[appConfig.numbers.ZERO_NUMBER] && setFirstTokenBlock(networks[appConfig.numbers.ZERO_NUMBER]);
        networks[appConfig.numbers.ONE_NUMBER] && setSecondTokenBlock(networks[appConfig.numbers.ONE_NUMBER]);
    }, [networks]);

    useEffect(() => {
        networks[appConfig.numbers.ZERO_NUMBER] && setSecondTokenBlock(isTokenReversed ? networks[appConfig.numbers.ZERO_NUMBER] : networks[appConfig.numbers.ONE_NUMBER]);
        networks[appConfig.numbers.ZERO_NUMBER] && setFirstTokenBlock(isTokenReversed ? networks[appConfig.numbers.ONE_NUMBER] : networks[appConfig.numbers.ZERO_NUMBER]);
    }, [isTokenReversed]);

    const hangleChangeAmount = (e: React.ChangeEvent<HTMLInputElement>) => {
        const numberPattern = /^([0-9]+)([\.,]{0,1})([0-9]*)$/g;
        const amount = e.target.value;
        if (amount === '') {
            setTokenAmount('');
        }

        if (amount.match(numberPattern)) {
            setTokenAmount(amount);
        }
    };

    const hangleChangeDestinationAddress = (e: React.ChangeEvent<HTMLInputElement>) => {
        const address = e.target.value;
        setDestinationAddress(address);
    };

    const estimateCurrentTransfer = async() => {
        if (!destinationAddress || !tokenAmount) {
            Notifications.emptySwapFields();
            return;
        }

        if (firstTokenBlock.type === NetworkTypes.EVM && !isMetamaskConnected) {
            Notifications.metamaskIsNotConnected();
            return;
        }

        if (firstTokenBlock.type === NetworkTypes.CASPER && !isCasperConnected) {
            Notifications.casperIsNotConnected();
            return;
        }

        if (firstTokenBlock.type === NetworkTypes.CASPER && !WalletAddressValidator.isEthAddressValid(destinationAddress)) {
            Notifications.walletAddressNotValid();
            return;
        }

        if (firstTokenBlock.type === NetworkTypes.EVM && !WalletAddressValidator.isCasperAccountHashValid(destinationAddress)) {
            Notifications.casperAccountHashIsNotValid();
            return;
        }

        // TODO: In feature will be deleted.
        if (firstTokenBlock.type === NetworkTypes.CASPER && !Number.isInteger(Number(tokenAmount))) {
            Notifications.tokenAmoutIsNotInteger();
            return;
        }

        const transferEstimateRequest = new TransferEstimateRequest(
            firstTokenBlock.name,
            secondTokenBlock.name,
            firstTokenBlock.id,
            tokenAmount,
        );
        try {
            await dispatch(estimateTransfer(transferEstimateRequest));
            setIsTransactionModalOpen(!isTransactionModalOpen);
        } catch (error) {
            Notifications.couldNotEstimateTransfer();
        }
    };

    return (
        <>
            <div className="swap">
                <div className="swap__header">
                    <h1 className="swap__header__title">Import</h1>
                    <button aria-label="Download wallet" className="swap__header__btn">
                        <img className="swap__header__btn__icon" src={download} alt="download" />
                    </button>
                </div>
                <div className="swap__tokens">
                    <div className="swap__tokens__item">
                        <div className="swap__tokens__item__logo">
                            <div className="swap__tokens__item__logo__image-wrapper">
                                <img src={eth} alt="eth" />
                            </div>
                        </div>
                        <div className="swap__tokens__item__select">
                            <span className="swap__tokens__item__selected">{firstTokenBlock.name}</span>
                            <div className="swap__tokens__item__select__triangle" />
                            <div className="swap__tokens__item__list"></div>
                        </div>
                    </div>
                    <div className="swap__tokens__swap-icon" onClick={() => setIsTokenReversed(!isTokenReversed)}>
                        <img src={swap} alt="swap" />
                    </div>
                    <div className="swap__tokens__item">
                        <div className="swap__tokens__item__logo">
                            <div className="swap__tokens__item__logo__image-wrapper">
                                <img src={eth} alt="eth" />
                            </div>
                        </div>
                        <div className="swap__tokens__item__select">
                            <span className="swap__tokens__item__selected">{secondTokenBlock.name}</span>
                            <div className="swap__tokens__item__select__triangle" />
                            <div className="swap__tokens__item__list"></div>
                        </div>
                    </div>
                </div>
                <div className="swap__tabs">
                    <span className={`swap__tabs__item ${swapTokenClassName}`} onClick={() => setSelectedSwapItem(SwapItems.Token)}>Token</span>
                    {/** TODO: onClick will be replaced to "() => setSelectedSwapItem(SwapItems.NFT)" */}
                    <span className={`swap__tabs__item ${swapNftClassName}`} onClick={() => ''}>NFT</span>
                </div>
                <div className="swap__token-list" onClick={() => setIsTokensModalOpen(true)}>
                    <div className="swap__token-list__circle"></div>
                    <span className="swap__token-list__selected-item">Asset (Price in USDT)</span>
                    <div className="swap__token-list__triangle" />
                </div>
                <div className="swap__amount" onChange={(e: React.ChangeEvent<HTMLInputElement>) => hangleChangeAmount(e)}>
                    <input className="swap__amount__value" type="string" placeholder="Amount" value={tokenAmount} aria-label="swap-amount" />
                    <button aria-label="Available balance" className="swap__amount__available-btn">Available</button>
                    <button aria-label="Max balance" className="swap__amount__max-btn">Max</button>
                </div>
                <div className="swap__comission">
                    <span className="swap__comission__label">Current fee is 5%</span>
                    <Info />
                    <span className="swap__comission__title">Commissions</span>
                </div>
                {
                    !isSelectedCheckbox &&
                    <div className="swap__destination" onChange={(e: React.ChangeEvent<HTMLInputElement>) => hangleChangeDestinationAddress(e)}>
                        <input className="swap__amount__value" type="text" placeholder="Destination" value={destinationAddress} aria-label="swap-destination" />
                    </div>
                }
                {
                    !getLocalStorageItem(LocalStorageKeys.isCasperConnected) && !getLocalStorageItem(LocalStorageKeys.isMetamaskConnected)
                    ?
                    <button aria-label="Connect wallet" className="swap__btn" onClick={() => setIsWalletsModalOpen(true)}>Connect wallet</button>
                    :
                    <button aria-label="Swap" className="swap__btn" onClick={estimateCurrentTransfer}>Swap</button>
                }
            </div>
            <TokensModal isOpen={isTokensModalOpen} onClose={() => setIsTokensModalOpen(false)} networkId={appConfig.numbers.ONE_NUMBER} />
            <TransactionModal
                isOpen={isTransactionModalOpen}
                onClose={() => setIsTransactionModalOpen(false)}
                fromTokenImg={eth}
                toTokenImg={eth}
                asset={activeSupportedToken.shortName}
                amount={tokenAmount}
                destination={destinationAddress}
                activeNetwork={firstTokenBlock.name}
            />
            <WalletsModal
                isOpen={isWalletsModalOpen}
                onClose={() => setIsWalletsModalOpen(false)}
                setWalletAddresses={setWalletAddresses}
                walletAddresses={walletAddresses}
            />
        </>
    );
};

export default Swap;
