import React, { Component } from 'react';
import PropTypes from 'prop-types';
import XTerm from 'xterm';

export default class Terminal extends Component {
    componentDidMount() {
        const terminal = new XTerm();

        terminal.open(document.getElementById('terminal'));
        terminal.write('test');
    }

    render() {
        return <div id="terminal" />;
    }
}
