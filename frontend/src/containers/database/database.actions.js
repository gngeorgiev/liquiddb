export const executeDatabaseCode = (code, db, print) => {
    const codeWrapper = `(function code() {
        return function (db) {
            return $code
        }; 
    }())`;

    return async dispatch => {
        const codeString = codeWrapper.replace('$code', code);
        try {
            // eslint-disable-next-line
            const codeFn = eval(codeString);
            const res = await Promise.resolve(codeFn(db));

            dispatch({
                type: 'EXECUTE_CODE',
                code: codeString,
                result: res
            });

            print(JSON.stringify(res));
        } catch (e) {
            dispatch({
                type: 'EXECUTE_CODE',
                code: codeString,
                result: e.toString()
            });

            print(e.toString());
            console.error(e);
        }
    };
};
