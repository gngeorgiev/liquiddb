import React, { Component } from 'react';
import PropTypes from 'prop-types';
import Card, { CardActions, CardContent, CardHeader } from 'material-ui/Card';
import Button from 'material-ui/Button';

export default class WidgetWrapper extends Component {
    static propTypes = {
        title: PropTypes.string.isRequired,
        routeName: PropTypes.string,
        actionPressed: PropTypes.func
    };

    render() {
        return (
            <div>
                <Card>
                    <CardHeader title={this.props.title} />
                    <CardContent>
                        {this.props.children}
                    </CardContent>
                    {this.props.routeName
                        ? <CardActions>
                              <Button
                                  onClick={() => this.props.actionPressed()}
                                  dense
                              >
                                  Go to {this.props.routeName} page
                              </Button>
                          </CardActions>
                        : null}
                </Card>
            </div>
        );
    }
}
