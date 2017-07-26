import React, { Component } from 'react';
import PropTypes from 'prop-types';
import Card, { CardActions, CardContent } from 'material-ui/Card';
import Button from 'material-ui/Button';
import Typography from 'material-ui/Typography';

export default class ConnectionsCount extends Component {
    static propTypes = {
        count: PropTypes.number.isRequired
    };

    render() {
        return (
            <div>
                <Card>
                    <CardContent>
                        <Typography type="title">
                            {this.props.count}
                        </Typography>
                    </CardContent>
                </Card>
            </div>
        );
    }
}
