import React, { Component } from 'react';
import { Route } from 'react-router-dom';
import { push } from 'react-router-redux';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import { LiquidDb } from 'liquiddb-javascript-driver/web';
import AppBar from 'material-ui/AppBar';
import Toolbar from 'material-ui/Toolbar';
import Typography from 'material-ui/Typography';
import IconButton from 'material-ui/IconButton';
import List, { ListItem, ListItemText } from 'material-ui/List';
import DashboardIcon from 'material-ui-icons/Dashboard';
import MenuIcon from 'material-ui-icons/Menu';
import StorageIcon from 'material-ui-icons/Storage';

import { Dashboard } from './containers/dashboard/dashboard.container';
import Database from './containers/database/database.container';

import './App.css';

LiquidDb.configureLogger({
    level: 'debug'
});

class App extends Component {
    routesMap = {
        '/': 'Dashboard',
        '/database': 'Database'
    };

    async componentDidMount() {
        this.db = await new LiquidDb().initialize();

        this.forceUpdate();
    }

    componentWillUnmount() {
        this.db.close();
    }

    render() {
        const { routesMap, db } = this;
        const { path } = this.props;

        if (db && db.ready) {
            return (
                <div>
                    <AppBar position="static">
                        <Toolbar>
                            <IconButton color="contrast" aria-label="Menu">
                                <MenuIcon />
                            </IconButton>
                            <Typography type="title" color="inherit">
                                {routesMap[path]}
                            </Typography>
                        </Toolbar>
                    </AppBar>

                    <div className="main-container">
                        <div className="menu">
                            <List>
                                <ListItem
                                    button
                                    onClick={() => this.props.goToDashboard()}
                                >
                                    <DashboardIcon />
                                    <ListItemText primary="Dashboard" />
                                </ListItem>
                                <ListItem
                                    button
                                    onClick={() => this.props.goToDatabase()}
                                >
                                    <StorageIcon />
                                    <ListItemText primary="Database" />
                                </ListItem>
                            </List>
                        </div>
                        <div className="main">
                            <Route
                                exact
                                path="/"
                                component={() => <Dashboard db={db} />}
                            />
                            <Route
                                exact
                                path="/database"
                                component={() =>
                                    <Database expand={true} db={db} />}
                            />
                        </div>
                    </div>
                </div>
            );
        }

        return <div>Loading....</div>;
    }
}

const mapStateToProps = state => ({
    path: state.routing && state.routing.location.pathname
});

const mapDispatchToProps = dispatch =>
    bindActionCreators(
        {
            goToDashboard: () => push('/'),
            goToDatabase: () => push('/database')
        },
        dispatch
    );

export default connect(mapStateToProps, mapDispatchToProps)(App);
