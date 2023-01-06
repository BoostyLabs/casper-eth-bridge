import renderer from 'react-test-renderer';
import configureStore from 'redux-mock-store';
import { cleanup } from "@testing-library/react";
import { useDispatch, useSelector, Provider } from "react-redux";

import TransactionsHistory from '@app/views/TransactionsHistory';

import { TransferEstimate, TransfersHistory } from '@/transfers';

describe("TransactionsHistory view", () => {
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
            .create(<Provider store={updatedStore}><TransactionsHistory /></Provider>).toJSON();
        expect(tree).toMatchSnapshot()
    })
});
