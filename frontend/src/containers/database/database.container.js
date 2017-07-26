import React, { Component } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import Terminal from 'terminal-in-react';
import 'terminal-in-react/lib/css/index.css';

import DatabaseViewer from '../../components/database-viewer/database-viewer.component';
import './database.container.css';

import { executeDatabaseCode } from './database.actions';

class Database extends Component {
    render() {
        return (
            <div className="database-container">
                <Terminal
                    msg="Write commands using the provided db object, a reference to a LiquidDb instance."
                    watchConsoleLogging={false}
                    startState="maximised"
                    commandPassThrough={(cmd, print) => {
                        this.props.executeDatabaseCode(
                            cmd.join(''),
                            this.props.db,
                            print
                        );
                    }}
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
