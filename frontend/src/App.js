import React, { Component } from 'react';
import './App.css';
import { Admin } from 'admin-on-rest';

import { Dashboard } from './Dashboard';

class App extends Component {
    render() {
        return (
            <Admin dashboard={Dashboard} restClient={() => Promise.resolve()} />
        );
    }
}

export default App;
