import React, { Component } from 'react';
import PropTypes from 'prop-types';
import DatabaseViewer from '../../components/database-viewer/database-viewer.component';

export class Dashboard extends Component {
    static propTypes = {
        db: PropTypes.any.isRequired
    };

    render() {
        return <DatabaseViewer expand={true} db={this.props.db} />;
    }
}
