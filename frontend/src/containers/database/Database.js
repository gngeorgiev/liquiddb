import React, { Component } from 'react';
import PropTypes from 'prop-types';
import { Database as DatabaseComponent } from '../../components/database/Database';
import AceEditor from 'react-ace';
import Button from 'material-ui/Button';
import CheckIcon from 'material-ui-icons/Check';
import CloseIcon from 'material-ui-icons/Close';
import * as moment from 'moment';

import 'brace/mode/javascript';
import 'brace/theme/monokai';
import './Database.css';

const codeWrapper = `(function code() {
    return async function (db) {
        $code
    }; 
}())`;

const codeSaveKey = '__editor_code';

export class Database extends Component {
    static propTypes = {
        db: PropTypes.any.isRequired
    };

    state = {
        lastSaved: 'never',
        editorDirty: false,
        code: ''
    };

    async runCode() {
        const { db } = this.props;
        const { code } = this.state;

        const codeString = codeWrapper.replace('$code', code);
        try {
            // eslint-disable-next-line
            const codeFn = eval(codeString);
            const res = await codeFn(db);
            console.log(res);
        } catch (e) {
            console.error(e);
        }
    }

    bindEditorCommands() {
        const { editor } = this.refs.editor;

        editor.commands.addCommand({
            name: 'save',
            bindKey: { win: 'Ctrl-S', mac: 'Cmd-S' },
            exec: () => {
                {
                    localStorage.setItem(
                        codeSaveKey,
                        JSON.stringify({
                            code: this.state.code,
                            timestamp: new Date().toISOString()
                        })
                    );

                    this.setState({ editorDirty: false });
                }
            }
        });

        editor.commands.addCommand({
            name: 'execute-code',
            bindKey: { win: 'Ctrl-Enter', mac: 'Cmd-Enter' },
            exec: () => this.runCode()
        });
    }

    loadLocalCode() {
        const savedCodeEntry = localStorage.getItem(codeSaveKey);
        if (savedCodeEntry) {
            const savedCode = JSON.parse(savedCodeEntry);
            this.setState({
                code: savedCode.code,
                lastSavedDate: savedCode.timestamp
            });
        }
    }

    initializeLastSavedInterval() {
        const intervalCallback = () => {
            const savedCodeEntry = JSON.parse(
                localStorage.getItem(codeSaveKey) || '{}'
            );
            if (!savedCodeEntry.timestamp) {
                return;
            }

            const date = new Date(savedCodeEntry.timestamp);
            if (isNaN(date.getTime())) {
                return;
            }

            const diff = moment().diff(date, 'seconds');
            this.setState({
                lastSaved: diff ? `${diff} seconds ago` : 'just now'
            });
        };

        this._lastSavedInterval = setInterval(intervalCallback, 5000);
        intervalCallback();
    }

    componentDidMount() {
        this.initializeLastSavedInterval();
        this.loadLocalCode();
        this.bindEditorCommands();
    }

    componentWillUnmount() {
        clearInterval(this._lastSavedInterval);
    }

    render() {
        return (
            <div className="database-container">
                <div>
                    {this.state.editorDirty
                        ? <CloseIcon className="save-icon" />
                        : <CheckIcon className="save-icon" />}
                    <span>
                        Last save: {this.state.lastSaved}
                    </span>
                </div>

                <AceEditor
                    mode="javascript"
                    theme="monokai"
                    name="editor"
                    height="200px"
                    onChange={code =>
                        this.setState({ editorDirty: true, code })}
                    value={this.state.code}
                    ref="editor"
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
