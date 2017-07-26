import { combineReducers } from 'redux';
import { routerReducer } from 'react-router-redux';
import databaseReducer from '../containers/database/database.reducers';
import * as rootReducers from './rootReducers';

export default combineReducers({
    ...rootReducers,
    routing: routerReducer,
    database: databaseReducer
});
