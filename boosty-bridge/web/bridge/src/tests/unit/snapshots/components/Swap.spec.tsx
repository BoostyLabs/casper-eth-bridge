import renderer from 'react-test-renderer';
import configureStore from 'redux-mock-store';
import { cleanup } from "@testing-library/react";
import { useDispatch, useSelector, Provider } from "react-redux";
import { BrowserRouter, Route, Routes } from 'react-router-dom';

import Swap from '@app/views/Swap';

import { Token } from '@/networks';
import { TransferEstimate, TransfersHistory } from '@/transfers';

describe("Swap view", () => {
    beforeEach(() => {
        useSelectorMock.mockClear();
        useDispatchMock.mockClear();
    });

    afterAll(() => {
        cleanup();
    });

    const reactRedux = { useDispatch, useSelector }
    const useDispatchMock = jest.spyOn(reactRedux, "useDispatch");
    const useSelectorMock = jest.spyOn(reactRedux, "useSelector");
    
    it('renders correctly', () => {
        const mockStore = configureStore();
        const initialState = {
            networksReducer: {
                networks: [],
                supportedTokens: [],
                activeSupportedToken: new Token()
            },
            transfersReducer: {
                history: new TransfersHistory(),
                transferEstimate: new TransferEstimate()
            }
        };
        let updatedStore = mockStore(initialState);

        const mockDispatch = jest.fn();
        useDispatchMock.mockReturnValue(mockDispatch);
        updatedStore.dispatch = mockDispatch;

        const tree = renderer
            .create(
                <Provider store={updatedStore}>
                    <BrowserRouter>
                        <Routes>
                            <Route path="/" element={<Swap />}/>
                        </Routes>
                    </BrowserRouter>
                </Provider>
            ).toJSON();
        expect(tree).toMatchSnapshot()
    })
});
