import React, { Component } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import CodeEditor from '../../components/code-editor/code-editor.component';
// import 'terminal-in-react/lib/css/index.css';

import DatabaseViewer from '../../components/database-viewer/database-viewer.component';
import './database.container.css';

import { executeDatabaseCode } from './database.actions';

class Database extends Component {
    render() {
        return (
            <div className="database-container">
                <CodeEditor
                    onExecute={code =>
                        this.props.executeDatabaseCode(
                            code,
                            this.props.db,
                            console.log.bind(console)
                        )}
                />

                <DatabaseViewer expand={true} data={this.props.data} />
            </div>
        );
    }
}

const mapProps = state => ({
    executionResult: state.database.executionResult,
    db: state.db,
    data: state.dbData
});

const mapDispatch = dispatch =>
    bindActionCreators(
        {
            executeDatabaseCode
        },
        dispatch
    );

export default connect(mapProps, mapDispatch)(Database);
