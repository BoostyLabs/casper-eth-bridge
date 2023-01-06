import './index.scss';

export const ThemeProvider: React.FC = ({ children }) => <div className="theme">
    <div className="theme__dividers">
        <div className="theme__divider" />
        <div className="theme__divider" />
        <div className="theme__divider" />
    </div>
    <div className="theme__left-balls">
        <div className="theme__left-ball-high" />
        <div className="theme__left-ball-medium" />
        <div className="theme__left-ball-low" />
    </div>
    <div className="theme__right-ball-wrapper">
        <div className="theme__right-ball" />
    </div>
    <div className="theme__bottom-ball-wrapper">
        <div className="theme__bottom-ball-high" />
        <div className="theme__bottom-ball-low" />
    </div>
    <div className="theme__dots">
        <div className="theme__dot" />
        <div className="theme__dot" />
        <div className="theme__dot" />
        <div className="theme__dot" />
        <div className="theme__dot" />
        <div className="theme__dot" />
        <div className="theme__dot" />
        <div className="theme__dot" />
        <div className="theme__dot" />
        <div className="theme__dot" />
        <div className="theme__dot" />
        <div className="theme__dot" />
        <div className="theme__dot" />
        <div className="theme__dot" />
        <div className="theme__dot" />
        <div className="theme__dot" />
        <div className="theme__dot" />
        <div className="theme__dot" />
    </div>
    <div className="theme__content">
        {children}
    </div>
</div>;
