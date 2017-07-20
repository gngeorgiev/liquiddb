import React, { Component } from 'react';
import JSONTree from 'react-json-tree';
import { LiquidDb } from 'liquiddb-javascript-driver/web';

LiquidDb.configureLogger({
    level: 'debug'
});

export class Dashboard extends Component {
    state = {
        data: {}
    };

    async componentDidMount() {
        this.db = await new LiquidDb().initialize();

        window.db = this.db;

        const refresh = async () => {
            const data = await this.db.value();

            this.setState({ data });
        };

        this._dataUnsubscribe = this.db.data(refresh);

        refresh();
    }

    componentWillUnmount() {
        this._dataUnsubscribe();
    }

    render() {
        return <JSONTree data={this.state.data} />;
    }
}
