import React, { Component } from 'react';
import './App.css';
import { Admin } from 'admin-on-rest';
import { LiquidDb } from 'liquiddb-javascript-driver/web';

import { Dashboard } from './Dashboard';

LiquidDb.configureLogger({
    level: 'debug'
});

class App extends Component {
    async componentDidMount() {
        this.db = await new LiquidDb().initialize();

        this.forceUpdate();
    }

    componentWillUnmount() {
        this.db.close();
    }

    render() {
        if (this.db && this.db.ready) {
            return (
                <Admin
                    dashboard={() => <Dashboard db={this.db} />}
                    restClient={() => Promise.resolve()}
                />
            );
        }

        return <div>Loading....</div>;
    }
}

export default App;
