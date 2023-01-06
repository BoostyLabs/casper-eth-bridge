import { Modal } from '@/app/components/common/Modal';

import './index.scss';

type CanceledTransactionModalProps = {
    isOpen: boolean;
    onClose: () => void;
    title: string;
};

export const CanceledTransaction: React.FC<CanceledTransactionModalProps> = ({
    isOpen,
    onClose,
    title,
}) =>
    <Modal isOpen={isOpen} onClose={onClose}>
        <div className="canceled-transaction__wrapper">
            <div className="canceled-transaction__header">
                <span className="canceled-transaction__header__title">
                    {title}
                </span>
            </div>
            <button className="canceled-transaction__body__save-btn" onClick={onClose}>Ok</button>
        </div>
    </Modal>;

