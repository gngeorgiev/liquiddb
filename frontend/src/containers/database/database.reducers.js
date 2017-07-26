export default (state = {}, action) => {
    switch (action.type) {
        case 'EXECUTE_CODE':
            return Object.assign({}, state, {
                executionResult: JSON.stringify(action.result)
            });

        default:
            return state;
    }
};
