import { combineReducers } from 'redux';
import { routerReducer } from 'react-router-redux';
import databaseReducer from '../containers/database/database.reducers';

export default combineReducers({
    routing: routerReducer,
    database: databaseReducer
});
