import { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { Modal } from '@/app/components/common/Modal';

import { RootState } from '@app/store';
import { Token } from '@/networks';
import { NotificationsPlugin as Notifications } from '@app/plugins/notifications';
import { getSupportedTokens, setActiveSupportedToken } from '@/app/store/actions/networks';

import { Cross } from '@static/images/svg';
import eth from '@static/images/tokens/eth.svg';

import './index.scss';

export type TokensModalProps = {
    isOpen: boolean;
    onClose: () => void;
    networkId: number;
};

export const TokensModal: React.FC<TokensModalProps> = ({ isOpen, onClose, networkId }) => {
    const { supportedTokens } = useSelector((state: RootState) => state.networksReducer);
    const dispatch = useDispatch();

    const changeActiveAsset = (token: Token) => {
        dispatch(setActiveSupportedToken(token));
        onClose();
    };

    useEffect(() => {
        (async() => {
            try {
                await dispatch(getSupportedTokens(networkId));
            } catch (error) {
                Notifications.couldNotGetSupportedTokens();
            }
        })();
    }, [networkId]);

    return (
        <Modal isOpen={isOpen} onClose={onClose}>
            <div className="tokens__wrapper">
                <div className="tokens__header">
                    <span className="tokens__header__title">Select Token</span>
                    <div onClick={onClose} className="tokens__header__close">
                        <Cross />
                    </div>
                </div>
                <input className="tokens__search" type="text" placeholder="Search" />
                <div className="tokens__list">
                    {
                        supportedTokens.length && supportedTokens.map((token: Token) =>
                            <div onClick={() => changeActiveAsset(token)} className="tokens__list__item" key={token.shortName}>
                                <img className="tokens__list__item__logo" src={eth} alt={token.shortName} />
                                <span className="tokens__list__item__short-name">{token.shortName}</span>
                                <span className="tokens__list__item__full-name">{token.longName}</span>
                                {/* TODO: amount to token entity not added. Uncoment after solution.
                                <span className="tokens__list__item__amount">${token.amount}</span>
                                */}
                            </div>
                        )
                    }
                </div>
            </div>
        </Modal>
    );
};
