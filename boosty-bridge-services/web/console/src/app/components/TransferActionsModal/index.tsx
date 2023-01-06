import { useRef } from 'react';

import { usePopupVisibility } from '@/app/hooks/usePopupVisibility';

import './index.scss';

type TransferActionsModalProps = {
    isCancelButtonVisible: boolean,
    onCancel: () => void,
    onCopy: () => void,
    setIsTransferActionsModalShown: (isTransferActionsModalShown: boolean) => void,
    // TODO: onSave
};

export const TransferActionsModal: React.FC<TransferActionsModalProps> = ({ isCancelButtonVisible, onCancel, onCopy, setIsTransferActionsModalShown }) => {
    const transferActionsModalRef = useRef(null);
    usePopupVisibility(transferActionsModalRef, setIsTransferActionsModalShown);
    const cancelButtonClassName: string = `transfer-actions__cancel${isCancelButtonVisible ? '-not-allowed' : ''}`;

    return <div ref={transferActionsModalRef} className="transfer-actions">
        <button className="transfer-actions__save">
            Save
        </button>
        <button onClick={onCopy} className="transfer-actions__copy">
            Copy
        </button>
        <button onClick={onCancel} className={cancelButtonClassName}>
            Cancel
        </button>
    </div>
};
