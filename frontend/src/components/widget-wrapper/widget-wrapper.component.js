import React, { Component } from 'react';
import PropTypes from 'prop-types';
import Card, { CardActions, CardContent, CardHeader } from 'material-ui/Card';
import Button from 'material-ui/Button';
import Typography from 'material-ui/Typography';

export default class WidgetWrapper extends Component {
    static propTypes = {
        title: PropTypes.string.isRequired,
        routeName: PropTypes.string.isRequired,
        actionPressed: PropTypes.func.isRequired,
        width: PropTypes.string
    };

    static defaultProps = {
        width: '33%'
    };

    render() {
        return (
            <div style={{ width: this.props.width }}>
                <Card>
                    <CardHeader title={this.props.title} />
                    <CardContent>
                        {this.props.children}
                    </CardContent>
                    <CardActions>
                        <Button
                            onClick={() => this.props.actionPressed()}
                            dense
                        >
                            Go to {this.props.routeName} page
                        </Button>
                    </CardActions>
                </Card>
            </div>
        );
    }
}
