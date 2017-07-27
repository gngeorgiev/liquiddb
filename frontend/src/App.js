import React, { Component } from 'react';
import { Route } from 'react-router-dom';
import { push } from 'react-router-redux';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import { LiquidDb, LiquidDbStats } from 'liquiddb-javascript-driver/web';
import AppBar from 'material-ui/AppBar';
import Toolbar from 'material-ui/Toolbar';
import Typography from 'material-ui/Typography';
import IconButton from 'material-ui/IconButton';
import List, { ListItem, ListItemText } from 'material-ui/List';
import DashboardIcon from 'material-ui-icons/Dashboard';
import MenuIcon from 'material-ui-icons/Menu';
import StorageIcon from 'material-ui-icons/Storage';
import Spinner from 'react-spinkit';

import Dashboard from './containers/dashboard/dashboard.container';
import Database from './containers/database/database.container';

import './App.css';

LiquidDb.configureLogger({
    level: 'debug'
});

class App extends Component {
    routesMap = {
        '/': 'Dashboard',
        '/database': 'Database',
        '/stats': 'Stats'
    };

    componentDidMount() {
        this.props.initializeDb();
        this.props.initializeDbStats();
    }

    componentWillUnmount() {
        this.props.db.close();
        this.props.dbStats.close();
    }

    render() {
        const { routesMap } = this;
        const { path, db } = this.props;

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
                                component={() => <Dashboard />}
                            />
                            <Route
                                exact
                                path="/database"
                                component={() => <Database />}
                            />
                        </div>
                    </div>
                </div>
            );
        }

        return (
            <div>
                <div>
                    <Spinner name="double-bounce" />
                </div>
                <div>Loading...</div>
            </div>
        );
    }
}

const mapStateToProps = state => ({
    path: state.routing && state.routing.location.pathname,
    db: state.db,
    dbStats: state.dbStats
});

const mapDispatchToProps = dispatch =>
    bindActionCreators(
        {
            goToDashboard: () => push('/'),
            goToDatabase: () => push('/database'),
            initializeDb: () => async dispatch => {
                const db = await new LiquidDb().connect();

                const refresh = async () => {
                    const data = await db.value();
                    dispatch({
                        type: 'DB_DATA',
                        data
                    });
                };

                dispatch({
                    type: 'INITIALIZE_DB',
                    db
                });

                db.data(refresh);
                refresh();
            },
            initializeDbStats: () => async dispatch => {
                const stats = await new LiquidDbStats().connect();

                stats.on('data', data => {
                    dispatch({
                        type: 'DB_STATS_DATA',
                        data
                    });
                });

                return dispatch({
                    type: 'INITIALIZE_DB_STATS',
                    stats
                });
            }
        },
        dispatch
    );

export default connect(mapStateToProps, mapDispatchToProps)(App);
