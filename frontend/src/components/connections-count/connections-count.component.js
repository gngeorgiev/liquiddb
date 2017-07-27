import React, { Component } from 'react';
import PropTypes from 'prop-types';
import Typography from 'material-ui/Typography';

export default class ConnectionsCount extends Component {
    static propTypes = {
        count: PropTypes.number
    };

    static defaultProps = {
        count: 0
    };

    render() {
        return (
            <div>
                <Typography type="title">
                    {this.props.count}
                </Typography>
            </div>
        );
    }
}
