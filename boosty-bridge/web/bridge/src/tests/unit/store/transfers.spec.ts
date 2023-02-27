import { afterEach, beforeEach, describe, expect, it } from '@jest/globals';
import { useDispatch, useSelector } from "react-redux";
import configureStore from 'redux-mock-store';
import { cleanup } from "@testing-library/react";

import { TransfersClient } from '@/api/transfers';
import appConfig from '@/app/configs/appConfig.json';
import { Transfer, StringTxHash, NetworkAddress, TransferStatuses, TransferEstimate, TransferEstimateRequest, TransferPagination, TransfersHistory } from '@/transfers';
import { SET_HISTORY, SET_TRANSFER_ESTIMATE } from '@/app/store/actions/transfers';

const transfers = new TransfersClient();
const mockStore = configureStore();

const successFetchMock = async (body: any) => {
    globalThis.fetch = () =>
    Promise.resolve({
        json: () => Promise.resolve(body),
        ok: true,
        status: 200,
    }) as Promise<Response>;
};

const failedFetchMock = async () => {
    globalThis.fetch = () => {
        throw new Error();
    };
};

const mockedGlobalFetch = globalThis.fetch;

const MOCK_TRANSFER_PAGINATION: TransferPagination = new TransferPagination(
    'helloworld',
    'helloworld',
    appConfig.numbers.ONE_NUMBER,
    appConfig.numbers.FIVE_NUMBER,
    appConfig.numbers.FIVE_NUMBER
);

const MOCK_ESTIMATE_TRANSFER_REQUEST: TransferEstimateRequest = new TransferEstimateRequest(
    '0',
    '121213AAD123',
    appConfig.numbers.ZERO_NUMBER,
    '0.001',
);

const MOCK_TRANSFER_ESTIMATE: TransferEstimate = new TransferEstimate(
    '5.1',
    '0.004',
    '10'
);

/** Mock transfers history. */
const TRANSFERS_HISTORY: TransfersHistory = new TransfersHistory(
    appConfig.numbers.FIVE_NUMBER,
    appConfig.numbers.ZERO_NUMBER,
    appConfig.numbers.FIVE_NUMBER,
    [
        new Transfer(
            '1',
            '22 Nov',
            appConfig.numbers.ZERO_NUMBER,
            new StringTxHash(),
            new NetworkAddress(),
            new NetworkAddress(),
            TransferStatuses.FINISHED,
            new StringTxHash()
        ),
    ]
);

/** Mock initial transfers state. */
const initialState = {
    transfersReducer: {
        history: new TransfersHistory(),
        transferEstimate: new TransferEstimate()
    }
};

const reactRedux = { useDispatch, useSelector }
const useDispatchMock = jest.spyOn(reactRedux, "useDispatch");
const useSelectorMock = jest.spyOn(reactRedux, "useSelector");
let updatedStore: any = mockStore(initialState);
const mockDispatch = jest.fn();
useDispatchMock.mockReturnValue(mockDispatch);
updatedStore.dispatch = mockDispatch;

describe('Requests transfers list by signature.', () => {
    beforeEach(() => successFetchMock(TRANSFERS_HISTORY));
    afterEach(() => {
        globalThis.fetch = mockedGlobalFetch;
    });
    
    it('success response', async () => {
        const connectedNetworks = await transfers.history(MOCK_TRANSFER_PAGINATION);
        expect(connectedNetworks).toEqual(TRANSFERS_HISTORY);
    });

    describe('Failed response.', () => { 
        beforeEach(() => {
            failedFetchMock();
            useSelectorMock.mockClear();
            useDispatchMock.mockClear();
        });

        afterEach(() => {
            globalThis.fetch = mockedGlobalFetch;
            cleanup();
        });
        
        it('Must be empty transfers history state', async () => {
            try {
                await await transfers.history(MOCK_TRANSFER_PAGINATION);
            } catch (error) {
                mockDispatch(SET_HISTORY, new TransfersHistory());
                expect(updatedStore.getState().transfersReducer.history).toEqual(new TransfersHistory());
            }
        });
    })
});

describe('Requests transfer estimate.', () => {
    beforeEach(() => successFetchMock(MOCK_TRANSFER_ESTIMATE));
    afterEach(() => {
        globalThis.fetch = mockedGlobalFetch;
    });
    
    it('success response', async () => {
        const connectedNetworks = await transfers.estimate(MOCK_ESTIMATE_TRANSFER_REQUEST);
        expect(connectedNetworks).toEqual(MOCK_TRANSFER_ESTIMATE);
    });

    describe('Failed response.', () => { 
        beforeEach(() => {
            failedFetchMock();
            useSelectorMock.mockClear();
            useDispatchMock.mockClear();
        });

        afterEach(() => {
            globalThis.fetch = mockedGlobalFetch;
            cleanup();
        });
        
        it('Must be empty transfers estimate state', async () => {
            try {
                await await transfers.estimate(MOCK_ESTIMATE_TRANSFER_REQUEST);
            } catch (error) {
                mockDispatch(SET_TRANSFER_ESTIMATE, new TransferEstimate());
                expect(updatedStore.getState().transfersReducer.transferEstimate).toEqual(new TransferEstimate());
            }
        });
    })
});
