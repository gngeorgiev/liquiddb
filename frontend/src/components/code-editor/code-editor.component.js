import React, { Component } from 'react';
import PropTypes from 'prop-types';
import AceEditor from 'react-ace';
import CheckIcon from 'material-ui-icons/Check';
import CloseIcon from 'material-ui-icons/Close';
import * as moment from 'moment';
import Button from 'material-ui/Button';

import 'brace/mode/javascript';
import 'brace/theme/monokai';
import './code-editor.component.css';

const codeSaveKey = '__editor_code';

export default class CodeEditor extends Component {
    static PropTypes = {
        onExecute: PropTypes.func.isRequired
    };

    state = {
        lastSaved: 'never',
        editorDirty: false,
        code: ''
    };

    bindEditorCommands() {
        const { editor } = this.refs.editor;

        editor.commands.addCommand({
            name: 'save',
            bindKey: { win: 'Ctrl-S', mac: 'Cmd-S' },
            exec: () => {
                localStorage.setItem(
                    codeSaveKey,
                    JSON.stringify({
                        code: this.state.code,
                        timestamp: new Date().toISOString()
                    })
                );

                this.setState({ editorDirty: false });
                this._updateLastSaved();
            }
        });

        editor.commands.addCommand({
            name: 'execute-code',
            bindKey: { win: 'Ctrl-Enter', mac: 'Cmd-Enter' },
            exec: () => this.props.onExecute(this.state.code)
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
        } else {
            this.setState({
                code: `function run(db) {
    return db.value();
}`
            });
        }
    }

    initializeLastSavedInterval() {
        this._updateLastSaved = () => {
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

        this._lastSavedInterval = setInterval(this._updateLastSaved, 5000);
        this._updateLastSaved();
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
            <div className="code-container">
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

                <Button
                    raised
                    color="primary"
                    onClick={() => this.props.onExecute(this.state.code)}
                >
                    Run Code
                </Button>
            </div>
        );
    }
}
