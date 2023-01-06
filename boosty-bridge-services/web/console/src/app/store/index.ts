import thunk from 'redux-thunk';
import { applyMiddleware, combineReducers, createStore } from 'redux';

import { networksReducer } from '@/app/store/reducers/networks';
import { transfersReducer } from '@/app/store/reducers/transfers';


const reducer = combineReducers({
    networksReducer,
    transfersReducer,
});

export const store = createStore(reducer, applyMiddleware(thunk));

export type RootState = ReturnType<typeof store.getState>;
