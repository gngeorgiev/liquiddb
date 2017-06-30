import React from 'react';
import { browserHistory, Router } from 'react-router';
import { Provider } from 'react-redux';
import PropTypes from 'prop-types';
import { Admin, Resource } from 'admin-on-rest';

import { Dashboard } from './Dashboard/dashboard.index';

class App extends React.Component {
    static propTypes = {
        store: PropTypes.object.isRequired,
        routes: PropTypes.object.isRequired
    };

    shouldComponentUpdate() {
        return false;
    }

    render() {
        return (
            <Provider store={this.props.store}>
                <Admin
                    restClient={() => Promise.resolve({})}
                    dashboard={Dashboard}
                />
                {/*<div style={{ height: '100%' }}>
          <Router history={browserHistory} children={this.props.routes} />
        </div>*/}
            </Provider>
        );
    }
}

export default App;
