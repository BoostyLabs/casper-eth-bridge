import { ToastContainer } from 'react-toastify';

import appConfig from '@/app/configs/appConfig.json';
import { PositionsOnPage } from '@app/plugins/notifications';

import './index.scss';

/** Custom notification wrapper component around toast notifications. */
export const Notification: React.FC = () => {
    /** Indicates if newest notifications shown in top of queue. */
    const IS_NEWEST_ON_TOP: boolean = false;

    return <ToastContainer
        position={PositionsOnPage.TOP_RIGHT}
        autoClose={appConfig.numbers.FIVE_THOUSAND_NUMBER}
        hideProgressBar
        newestOnTop={IS_NEWEST_ON_TOP}
        pauseOnFocusLoss
        pauseOnHover
    />;
};
