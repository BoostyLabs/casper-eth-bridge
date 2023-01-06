import { useState } from 'react';

import { TransferActionsModal } from '@/app/components/TransferActionsModal';

import appConfig from '@/app/configs/appConfig.json';
import { getDayAndMonth, getHoursAndMinutes } from '@/app/internal/time';
import { Transfer, TransferStatuses } from '@/transfers';

import copy from '@static/images/common/copy.svg';

type TransferItemProps = {
    cancelTransfer: (id: number) => void,
    copyWalletAddress: (walletAddress: string) => void,
    encodeTransferStatus: (transferStatus: TransferStatuses) => number,
    transfer: Transfer,
};

export const TransferItem: React.FC<TransferItemProps> = ({ cancelTransfer, copyWalletAddress, encodeTransferStatus, transfer }) => {
    const [isTransferActionsModalShown, setIsTransferActionsModalShown] = useState<boolean>(false);
    const getWalletAddressLabel = (walletAddress: string) => `${walletAddress.slice(appConfig.numbers.ZERO_NUMBER, appConfig.numbers.FOUR_NUMBER)}...${walletAddress.slice(-appConfig.numbers.FOUR_NUMBER)}`;
    const isCancelButtonVisible: boolean = encodeTransferStatus(transfer.status) === TransferStatuses.CANCELED;
    /** Indicates if recipient is casper account. */
    const IS_RECIPIENT_CASPER_ACCOUNT: boolean = transfer.recipient.address.includes(appConfig.strings.CASPER_ACCOUNT_HASH_LABEL);

    const changeTransferActionsModalVisibility = () => {
        setIsTransferActionsModalShown(!isTransferActionsModalShown);
    };

    /** Returns recipient address depends on account. */
    const recipientAddress = () => {
        if (IS_RECIPIENT_CASPER_ACCOUNT) {
            return transfer.recipient.address.replace(appConfig.strings.CASPER_ACCOUNT_HASH_LABEL, '');
        }
        return transfer.recipient.address;
    };

    const onCancel = () => {
        if (isCancelButtonVisible) {
            return;
        }
        cancelTransfer(transfer.id);
        changeTransferActionsModalVisibility();
    };

    const onCopy = () => {
        copyWalletAddress(recipientAddress());
        changeTransferActionsModalVisibility();
    };

    return <li className="transactions-history__list__item">
        <span className="transactions-history__list__item__value">
            {transfer.id}
        </span>
        <span className="transactions-history__list__item__value">
            {transfer.sender.networkName}
        </span>
        <span className="transactions-history__list__item__value">
            {transfer.recipient.networkName}
        </span>
        <span
            className="transactions-history__list__item__destination"
            onClick={() => copyWalletAddress(recipientAddress())}
        >
            {getWalletAddressLabel(transfer.recipient.address)}
            <img
                alt="copy"
                className="transactions-history__list__item__destination__copy"
                src={copy}
            />
        </span>
        <span className="transactions-history__list__item__value">
            {transfer.amount}
        </span>
        <span className="transactions-history__list__item__value">
            {getHoursAndMinutes(transfer.createdAt)}
        </span>
        <span className="transactions-history__list__item__value">
            {getDayAndMonth(transfer.createdAt)}
        </span>
        <span className="transactions-history__list__item__value">
            {TransferStatuses[encodeTransferStatus(transfer.status)]}
        </span>
        <span className="transactions-history__list__item__value">
            <button onClick={changeTransferActionsModalVisibility} aria-label="More actions" className="transactions-history__list__item__value__action">
                ...
            </button>
        </span>
        {
            isTransferActionsModalShown &&
                <TransferActionsModal
                    isCancelButtonVisible={isCancelButtonVisible}
                    onCancel={onCancel}
                    onCopy={onCopy}
                    setIsTransferActionsModalShown={setIsTransferActionsModalShown}
                />
        }
    </li>
};
