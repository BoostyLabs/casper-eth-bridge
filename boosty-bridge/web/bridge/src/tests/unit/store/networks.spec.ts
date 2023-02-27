import { afterEach, beforeEach, describe, expect, it } from '@jest/globals';
import { useDispatch, useSelector } from "react-redux";
import configureStore from 'redux-mock-store';
import { cleanup } from "@testing-library/react";

import { NetworksClient } from '@/api/networks';
import appConfig from '@/app/configs/appConfig.json';
import { Network, NetworkNames, NetworkTypes, Token } from '@/networks';
import { SET_CONNECTED_NETWORKS, SET_SUPPORTED_TOKENS } from '@/app/store/actions/networks';

const networks = new NetworksClient();
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

/** Mock connected networks list. */
const CONNECTED_NETWORKS: Network[] = [
    new Network(
        appConfig.numbers.ZERO_NUMBER,
        NetworkNames.CASPER_TEST,
        NetworkTypes.CASPER,
        true,
    )
];

/** Mock supported tokens list. */
const SUPPORTED_TOKENS: Token[] = [
    new Token(
        appConfig.numbers.ZERO_NUMBER,
        'TST TOKEN',
        'TST',
        []
    ),
    new Token(
        appConfig.numbers.ONE_NUMBER,
        'TEST TOKEN',
        'TEST',
        [],
    ),
];

/** Mock initial networks state. */
const initialState = {
    networksReducer: {
        networks: [],
        supportedTokens: [],
        activeSupportedToken: new Token()
    }
};

const reactRedux = { useDispatch, useSelector }
const useDispatchMock = jest.spyOn(reactRedux, "useDispatch");
const useSelectorMock = jest.spyOn(reactRedux, "useSelector");
let updatedStore: any = mockStore(initialState);
const mockDispatch = jest.fn();
useDispatchMock.mockReturnValue(mockDispatch);
updatedStore.dispatch = mockDispatch;

describe('Requests connected networks list.', () => {
    beforeEach(() => {
        successFetchMock(CONNECTED_NETWORKS);
    });

    afterEach(() => {
        globalThis.fetch = mockedGlobalFetch;
    });
    
    it('Success response', async () => {
        const connectedNetworks = await networks.connected();
        expect(connectedNetworks).toEqual(CONNECTED_NETWORKS);
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
        
        it('Must be empty networks list', async () => {
            try {
                await networks.connected();
            } catch (error) {
                mockDispatch(SET_CONNECTED_NETWORKS, []);
                expect(updatedStore.getState().networksReducer.networks).toEqual([]);
            }
        });
    })
});

describe('Requests supported tokens list.', () => {
    beforeEach(() => successFetchMock(SUPPORTED_TOKENS));
    afterEach(() => {
        globalThis.fetch = mockedGlobalFetch;
    });

    it('Success response.', async () => {
        const connectedNetworks = await networks.supportedTokens(1);
        expect(connectedNetworks).toEqual(SUPPORTED_TOKENS);
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
        
        it('Must be empty supported tokens list', async () => {
            try {
                await networks.supportedTokens(appConfig.numbers.ZERO_NUMBER);
            } catch (error) {
                mockDispatch(SET_SUPPORTED_TOKENS, []);
                expect(updatedStore.getState().networksReducer.supportedTokens).toEqual([]);
            }
        });
    })
});
