import { useEffect } from 'react';

/** Hook allows control popup visibility. */
export const usePopupVisibility = (ref: React.MutableRefObject<null | HTMLElement>, setIsPopupVisible: (isPopupVisible: boolean) => void) => {
    useEffect(() => {
        /** Handles outside click. */
        function handleClickOutside(e: any) {
            if (ref.current && !ref.current.contains(e.target)) {
                setIsPopupVisible(false);
            }
        };

        /** Binds the event listener. */
        document.addEventListener('mousedown', handleClickOutside);

        return () => {
            /** Unbinds the event listener on clean up. */
            document.removeEventListener('mousedown', handleClickOutside);
        };
    }, [ref]);
};
