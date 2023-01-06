import './index.scss';

type SearchProps = {
    changeValue: (e: React.ChangeEvent<HTMLInputElement>) => void;
    value: string;
};

export const Search: React.FC<SearchProps> = ({ changeValue, value }) =>
    <div className="search">
        <input
            value={value}
            className="search__input"
            placeholder="Search"
            onChange={changeValue}
        />
        <button aria-label="Search" className="search__button">
            Search
        </button>
    </div>;
