import React, { Component } from 'react';
import PropTypes from 'prop-types';
import { Database } from '../../components/database/Database';

export class Dashboard extends Component {
    static propTypes = {
        db: PropTypes.any.isRequired
    };

    render() {
        return <Database expand={false} db={this.props.db} />;
    }
}
