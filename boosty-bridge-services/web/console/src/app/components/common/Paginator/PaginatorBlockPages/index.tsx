import { PaginatorEvents } from '@/app/components/common/Paginator';

type PaginatorBlockPagesProps = {
    blockPages: number[];
    onPageChange: (event: PaginatorEvents, pageNumber?: number) => void;
    currentPage: number;
};

export const PaginatorBlockPages: React.FC<PaginatorBlockPagesProps> = ({ blockPages, onPageChange, currentPage }) => {
    const activePageItemClassNameLabel = (page: number) => `paginator__pages__item${currentPage === page ? '-active' : ''}`;

    return <ul className="paginator__pages">
        {
            blockPages.map((page, index) =>
                <li
                    className={activePageItemClassNameLabel(page)}
                    key={index}
                    onClick={() => onPageChange(PaginatorEvents.change, page)}
                >
                    {page}
                </li>
            )
        }
    </ul>;
};
