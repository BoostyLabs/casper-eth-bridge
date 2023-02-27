import { useEffect, useState } from 'react';

import { PaginatorBlockPages } from '@/app/components/common/Paginator/PaginatorBlockPages';

import appConfig from '@/app/configs/appConfig.json';

import './index.scss';

export enum PaginatorEvents {
    change = 'change',
    next = 'next',
    previous = 'previous',
};

type PaginatorProps = {
    itemsCount: number;
    changeOffset: (offset: number) => void;
};

export const Paginator: React.FC<PaginatorProps> = ({ itemsCount, changeOffset }) => {
    const [currentPage, setCurrentPage] = useState<number>(appConfig.numbers.ONE_NUMBER);

    /** Splits the page into 3 blocks that can be needed to separate page numbers. */
    const [firstBlockPages, setFirstBlockPages] = useState<number[]>([]);
    const [middleBlockPages, setMiddleBlockPages] = useState<number[]>([]);
    const [lastBlockPages, setLastBlockPages] = useState<number[]>([]);

    const CARDS_ON_PAGE: number = appConfig.numbers.FIVE_NUMBER;
    const MAX_PAGES_PER_BLOCK: number = appConfig.numbers.FOUR_NUMBER;
    const MAX_PAGES_OFF_BLOCKS: number = appConfig.numbers.FIVE_NUMBER;
    const FIRST_PAGE_INDEX: number = appConfig.numbers.ZERO_NUMBER;
    const SECOND_PAGE_INDEX: number = appConfig.numbers.ONE_NUMBER;
    const FIRST_PAGE_INDEX_FROM_END: number = -appConfig.numbers.ONE_NUMBER;
    const NEG_STEP_FROM_CURRENT_PAGE: number = -appConfig.numbers.TWO_NUMBER;
    const POS_STEP_FROM_CURRENT_PAGE: number = appConfig.numbers.ONE_NUMBER;
    const FIRST_PAGE: number = 1;

    const pages: number[] = [];
    const pagesCount = Math.ceil(itemsCount / CARDS_ON_PAGE);

    for (let i = appConfig.numbers.ONE_NUMBER; i <= pagesCount; i++) {
        pages.push(i);
    }

    /** Indicates visibility of dots after first pages block. */
    const isFirstDotsShown: boolean = middleBlockPages.length <= MAX_PAGES_PER_BLOCK && pages.length > MAX_PAGES_OFF_BLOCKS;
    /** Indicates visibility of dots after middle pages block. */
    const isSecondDotsShown: boolean = !!middleBlockPages.length;
    /** Indicates in which block is current page. */
    const isCurrentInFirstBlock: boolean = currentPage < MAX_PAGES_PER_BLOCK;
    const isCurrentInLastBlock: boolean = pages.length - currentPage < MAX_PAGES_PER_BLOCK - SECOND_PAGE_INDEX;
    /** Change page blocks reorganization depends on current page. */
    const isOneBlockRequired: boolean = pages.length <= MAX_PAGES_OFF_BLOCKS;
    /** Indicates if current page is first page or last page. */
    const isFirstPageSelected: boolean = currentPage === FIRST_PAGE;
    const isLastPageSelected: boolean = currentPage === pagesCount;

    const previousPageClassNameLabel: string = `paginator__previous${isFirstPageSelected ? '-not-active' : ''}`;
    const nextPageClassNameLabel: string = `paginator__next${isLastPageSelected ? '-not-active' : ''}`;

    /** Sets block pages depend on current page. */
    const setBlocksIfCurrentInFirstBlock = () => {
        setFirstBlockPages(pages.slice(FIRST_PAGE_INDEX, MAX_PAGES_PER_BLOCK));
        setMiddleBlockPages([]);
        setLastBlockPages(pages.slice(FIRST_PAGE_INDEX_FROM_END));
    };
    const setBlocksIfCurrentInMiddleBlock = () => {
        setFirstBlockPages(pages.slice(FIRST_PAGE_INDEX, SECOND_PAGE_INDEX));
        setMiddleBlockPages(
            pages.slice(
                currentPage + NEG_STEP_FROM_CURRENT_PAGE,
                currentPage + POS_STEP_FROM_CURRENT_PAGE
            )
        );
        setLastBlockPages(pages.slice(FIRST_PAGE_INDEX_FROM_END));
    };
    const setBlocksIfCurrentInLastBlock = () => {
        setFirstBlockPages(pages.slice(FIRST_PAGE_INDEX, SECOND_PAGE_INDEX));
        setMiddleBlockPages([]);
        setLastBlockPages(pages.slice(-MAX_PAGES_PER_BLOCK));
    };

    const reorganizePagesBlock = () => {
        if (isOneBlockRequired) {
            return;
        }

        if (isCurrentInFirstBlock) {
            setBlocksIfCurrentInFirstBlock();
            return;
        }

        if (!isCurrentInFirstBlock && !isCurrentInLastBlock) {
            setBlocksIfCurrentInMiddleBlock();
            return;
        }

        if (isCurrentInLastBlock) {
            setBlocksIfCurrentInLastBlock();
        }
    };

    /*
     * Indicates if dots delimiter is needed to separate page numbers.
     */
    const populatePages = () => {
        if (!pages.length) {
            return;
        }
        if (isOneBlockRequired) {
            setFirstBlockPages(pages.slice());
            setMiddleBlockPages([]);
            setLastBlockPages([]);

            return;
        }
        reorganizePagesBlock();
    };

    /** Changes current page and sets pages block. */
    const onPageChange = (event: PaginatorEvents, pageNumber: number = currentPage): void => {
        switch (event) {
        case PaginatorEvents.next:
            if (pageNumber < pages.length) {
                setCurrentPage(pageNumber + appConfig.numbers.ONE_NUMBER);
            }
            populatePages();

            return;
        case PaginatorEvents.previous:
            if (pageNumber > SECOND_PAGE_INDEX) {
                setCurrentPage(pageNumber - appConfig.numbers.ONE_NUMBER);
            }
            populatePages();

            return;
        case PaginatorEvents.change:
            setCurrentPage(pageNumber);
            populatePages();

            return;
        default:
            populatePages();
        }
    };

    useEffect(() => {
        populatePages();
        changeOffset((currentPage - appConfig.numbers.ONE_NUMBER) * appConfig.numbers.FIVE_NUMBER);
    }, [currentPage, pagesCount]);

    return (
        <section className="paginator">
            <div className="paginator__wrapper">
                <a
                    className={previousPageClassNameLabel}
                    onClick={() => onPageChange(PaginatorEvents.previous)}
                >
                    <span className="paginator__previous__label">
                        &#60;
                    </span>
                </a>
                {
                    Boolean(firstBlockPages.length) &&
                        <PaginatorBlockPages
                            blockPages={firstBlockPages}
                            onPageChange={onPageChange}
                            currentPage={currentPage}
                        />
                }
                {isFirstDotsShown &&
                    <span className="paginator__pages__dots">
                        ...
                    </span>
                }
                {
                    Boolean(middleBlockPages.length) &&
                        <PaginatorBlockPages
                            blockPages={middleBlockPages}
                            onPageChange={onPageChange}
                            currentPage={currentPage}
                        />
                }
                {isSecondDotsShown &&
                    <span className="paginator__pages__dots">
                        ...
                    </span>
                }
                {
                    Boolean(lastBlockPages.length) &&
                        <PaginatorBlockPages
                            blockPages={lastBlockPages}
                            onPageChange={onPageChange}
                            currentPage={currentPage}
                        />
                }
                <a
                    className={nextPageClassNameLabel}
                    onClick={() => onPageChange(PaginatorEvents.next)}
                >
                    <span className="paginator__next__label">
                        &#62;
                    </span>
                </a>
            </div>
        </section>
    );
};
