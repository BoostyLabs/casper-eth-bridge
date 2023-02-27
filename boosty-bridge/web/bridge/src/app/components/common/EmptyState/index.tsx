import emptyState from '@/app/static/images/common/empty-state.svg';

import './index.scss';

export const EmptyState: React.FC = () => {
    return <div className="empty">
        <img
            src={emptyState}
            alt="empty state"
            className="empty__image"
        />
        <h1 className="empty__title">No Data Yet</h1>
        <span className="empty__description">
            There are no transactions associated with the connected wallet yet. Submit a transaction or connect another wallet.
        </span>
    </div>
};
