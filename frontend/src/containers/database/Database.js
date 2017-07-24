import React, { Component } from 'react';
import PropTypes from 'prop-types';
import { Database as DatabaseComponent } from '../../components/database/Database';
import AceEditor from 'react-ace';
import Button from 'material-ui/Button';

import 'brace/mode/javascript';
import 'brace/theme/monokai';
import './Database.css';

const codeWrapper = `(function code() {
    return async function (db) {
        $code
    }; 
}())`;

export class Database extends Component {
    static propTypes = {
        db: PropTypes.any.isRequired
    };

    //using state for the ace editor doesn't work very well
    code = '';

    async runCode() {
        const { code } = this;
        const { db } = this.props;

        const codeString = codeWrapper.replace('$code', code);
        try {
            const codeFn = eval(codeString);
            const res = await codeFn(db);
            console.log(res);
        } catch (e) {
            console.error(e);
        }
    }

    render() {
        return (
            <div className="database-container">
                <AceEditor
                    mode="javascript"
                    theme="monokai"
                    name="editor"
                    height="200px"
                    onChange={code => (this.code = code)}
                    width="100%"
                />

                <Button raised color="primary" onClick={() => this.runCode()}>
                    Run Code
                </Button>

                <DatabaseComponent expand={true} db={this.props.db} />
            </div>
        );
    }
}
