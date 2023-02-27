import './index.scss';

export type ModalProps = {
    isOpen: boolean;
    onClose: () => void;
};

export const Modal: React.FC<ModalProps> = ({ children, isOpen, onClose }) => {
    document.body.classList.remove('modal-open');

    if (!isOpen) { return null; }
    document.body.classList.add('modal-open');

    return (
        <div className="modal-window">
            <div className="modal-window__clickaway" onClick={onClose}/>
            <div className="modal-window__body">
                <div className="modal-window__content">
                    {children}
                </div>
            </div>
        </div>
    );
};
