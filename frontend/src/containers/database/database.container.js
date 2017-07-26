import React, { Component } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import PropTypes from 'prop-types';
import Terminal from 'terminal-in-react';
import 'terminal-in-react/lib/css/index.css';

import DatabaseViewer from '../../components/database-viewer/database-viewer.component';
// import CodeEditor from '../../components/code-editor/code-editor.component';
import './database.container.css';

import { executeDatabaseCode } from './database.actions';

class Database extends Component {
    static propTypes = {
        db: PropTypes.any.isRequired
    };

    render() {
        return (
            <div className="database-container">
                {/* <CodeEditor
                    onExecute={code =>
                        this.props.executeDatabaseCode(code, this.props.db)}
                /> */}

                <Terminal
                    msg="Write commands using the provided db object, a reference to a LiquidDb instance."
                    watchConsoleLogging={false}
                    startState="maximised"
                    commandPassThrough={(cmd, print) => {
                        this.props.executeDatabaseCode(
                            cmd,
                            this.props.db,
                            print
                        );
                    }}
                />

                {/* <div>
                    {this.props.executionResult}
                </div> */}

                <DatabaseViewer expand={true} db={this.props.db} />
            </div>
        );
    }
}

const mapProps = state => ({
    executionResult: state.database.executionResult
});

const mapDispatch = dispatch =>
    bindActionCreators(
        {
            executeDatabaseCode
        },
        dispatch
    );

export default connect(mapProps, mapDispatch)(Database);
