import React, { Component } from 'react';
import JSONTree from 'react-json-tree';
import PropTypes from 'prop-types';

export default class DatabaseViewer extends Component {
    static propTypes = {
        data: PropTypes.any.isRequired,
        expand: PropTypes.bool
    };

    render() {
        return (
            <JSONTree
                data={this.props.data}
                shouldExpandNode={() => this.props.expand}
            />
        );
    }
}
