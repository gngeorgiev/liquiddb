export const executeDatabaseCode = (code, db, print) => {
    const codeWrapper = `(function code() {
    return async function (db) {
        return $code
    }; 
}())`;

    return async dispatch => {
        const codeString = codeWrapper.replace('$code', code);
        try {
            // eslint-disable-next-line
            const codeFn = eval(codeString);
            const resValue = await codeFn(db);
            const res = await Promise.resolve(resValue);
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
        }
    };
};
