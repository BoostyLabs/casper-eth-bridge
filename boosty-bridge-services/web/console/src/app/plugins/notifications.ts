import { toast } from 'react-toastify';

import noficationsMessages from '@/app/configs/noficationsMessages.json';

/** Notifications themes. Contains colored, light and dark themes. */
export enum Themes {
    colored = 'colored',
    dark = 'dark',
    light = 'light',
};

/** Notifications types. I.e, error, info, success, warning. */
export enum DesignTypes {
    error = 'error',
    info = 'info',
    success = 'success',
    warning = 'warning',
};

/** Notifications position on page. */
export enum PositionsOnPage {
    BOTTOM_CENTER = 'bottom-center',
    BOTTOM_LEFT = 'bottom-left',
    BOTTOM_RIGHT = 'bottom-right',
    TOP_CENTER = 'top-center',
    TOP_LEFT = 'top-left',
    TOP_RIGHT = 'top-right',
};

/** Defines notifications plugin with message, toast type and theme. */
export class NotificationsPlugin {
    /** Notifies user. As default type uses error type, and default theme is colored. */
    static notify(
        message: string,
        type: DesignTypes = DesignTypes.error,
        theme: Themes = Themes.colored
    ) {
        toast[type](
            message,
            {
                position: toast.POSITION.TOP_RIGHT,
                theme,
            }
        );
    };

    /** Notifies that casper wallet isn't connected. */
    static casperIsNotConnected() {
        this.notify(noficationsMessages.casperIsNotConnected);
    };

    /** Notifies that message was copied. */
    static copied() {
        this.notify(noficationsMessages.copied, DesignTypes.info);
    };

    /** Notifies that could not cancel transfer. */
    static couldNotCancelTransfer() {
        this.notify(noficationsMessages.couldNotCancelTransfer);
    };

    /** Notifies that could not authenticate via MetaMask. */
    static couldNotAuthenficateViaMetaMask() {
        this.notify(noficationsMessages.couldNotAuthenficateViaMetaMask);
    };

    /** Notifies that could not connect via MetaMask. */
    static couldNotConnectViaMetaMask() {
        this.notify(noficationsMessages.couldNotConnectViaMetaMask);
    };

    /** Notifies that could not connect via Phantom. */
    static couldNotConnectViaPhantom() {
        this.notify(noficationsMessages.couldNotConnectViaPhantom);
    };

    /** Notififes that could not get connected networks. */
    static couldNotGetConnectedNetworks() {
        this.notify(noficationsMessages.couldNotGetConnectedNetworks);
    };

    /** Notifies that could not get supported tokens. */
    static couldNotGetSupportedTokens() {
        this.notify(noficationsMessages.couldNotGetSupportedTokens);
    };

    /** Notififes that could not get transfers history. */
    static couldNotGetTransfersHistory() {
        this.notify(noficationsMessages.couldNotGetTransfersHistory);
    };

    /** Notifies that could not get casper wallet address. */
    static couldNotGetCasperWalletAddres() {
        this.notify(noficationsMessages.couldNotGetCasperWalletAddres)
    };

    /** Notifies that could not get metamask wallet address. */
    static couldNotGetMetamaskWalletAddres() {
        this.notify(noficationsMessages.couldNotGetMetamaskWalletAddres)
    };

    /** Notifies that could not get Phantom wallet public key. */
    static couldNotGetPhantomWalletPublicKey() {
        this.notify(noficationsMessages.couldNotGetPhantomWalletPublicKey)
    };

    /** Notififes that could not estimate trasnfer. */
    static couldNotEstimateTransfer() {
        this.notify(noficationsMessages.couldNotEstimateTransfer);
    };

    /** Notifies that could not send transaction via Casper wallet. */
    static couldNotSendTransactionViaCasperWallet() {
        this.notify(noficationsMessages.couldNotSendTransactionViaCasperWallet);
    };

    /** Notifies that could not send transaction via MetaMask wallet. */
    static couldNotSendTransactionViaMetaMaskWallet() {
        this.notify(noficationsMessages.couldNotSendTransactionViaMetaMaskWallet);
    };

    /** Notifies that swap fields must be not empty. */
    static emptySwapFields() {
        this.notify(noficationsMessages.emptySwapFields);
    };

    /** Notifies that metamask wallet isn't connected. */
    static metamaskIsNotConnected() {
        this.notify(noficationsMessages.metamaskIsNotConnected);
    };

    /** Notifies that could not connect Casper. */
    static connectCasperError() {
        this.notify(noficationsMessages.connectCasperError)
    };

    /** Notifies that action was success. */
    static success() {
        this.notify(noficationsMessages.success, DesignTypes.success);
    };

    /** Notifies that transaction via MetaMask was canceled. */
    static transactionViaMetaMaskWasCanceled() {
        this.notify(noficationsMessages.transactionViaMetaMaskWasCanceled);
    };

    /** Notifies that action was success transaction. */
    static transactionSuccess() {
        this.notify(noficationsMessages.transactionSuccess, DesignTypes.success);
    };

    /** Notifies that transfer successfully canceled. */
    static transferSuccessfullyCanceled() {
        this.notify(noficationsMessages.transferSuccessfullyCanceled, DesignTypes.success);
    };

    /** Notifies that need to unlock Casper Signer. */
    static unlockCasperSigner() {
        this.notify(noficationsMessages.unlockCasperSigner);
    };

    /** Notifies that wallet successfully connected. */
    static walletSuccessfullyConnected() {
        this.notify(noficationsMessages.walletSuccessfullyConnected, DesignTypes.success);
    };

    /** Notifies that wallet addres is not valid. */
    static walletAddressNotValid() {
        this.notify(noficationsMessages.walletAddressNotValid)
    }

    /** Notifies that casper account hash isn't valid */
    static casperAccountHashIsNotValid() {
        this.notify(noficationsMessages.casperAccountHashIsNotValid);
    };

    // TODO: For MVP (in feature will be deleted)
    /** Notifies that token amount isn't integer */
    static tokenAmoutIsNotInteger() {
        this.notify(noficationsMessages.tokenAmoutIsNotInteger);
    };
};
