import React, { Component } from 'react';
import PropTypes from 'prop-types';
import Card, { CardActions, CardContent, CardHeader } from 'material-ui/Card';
import Button from 'material-ui/Button';

export default class WidgetWrapper extends Component {
    static propTypes = {
        title: PropTypes.string.isRequired,
        routeName: PropTypes.string,
        actionPressed: PropTypes.func,
        style: PropTypes.any
    };

    render() {
        const {
            title,
            children,
            routeName,
            actionPressed,
            ...props
        } = this.props;

        return (
            <div {...props}>
                <Card>
                    <CardHeader title={title} />
                    <CardContent>
                        {children}
                    </CardContent>
                    {routeName
                        ? <CardActions>
                              <Button onClick={() => actionPressed()} dense>
                                  Go to {routeName} page
                              </Button>
                          </CardActions>
                        : null}
                </Card>
            </div>
        );
    }
}
