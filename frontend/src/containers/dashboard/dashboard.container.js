import React, { Component } from 'react';
import { connect } from 'react-redux';
import { push } from 'react-router-redux';
import { bindActionCreators } from 'redux';

import DatabaseViewer from '../../components/database-viewer/database-viewer.component';
import ConnectionsCount from '../../components/connections-count/connections-count.component';
import WidgetWrapper from '../../components/widget-wrapper/widget-wrapper.component';

import './dashboard.container.css';

class DashboardContainer extends Component {
    render() {
        return (
            <div className="container">
                <div className="container-row">
                    <WidgetWrapper
                        title="Connections"
                        routeName="Stats"
                        actionPressed={() => this.props.goToStatsPage()}
                    >
                        <ConnectionsCount
                            count={this.props.dbStats.connectionsCount}
                        />
                    </WidgetWrapper>
                </div>
                <div className="container-row">
                    <WidgetWrapper
                        title="Database View"
                        routeName="Database"
                        width="100%"
                        actionPressed={() => this.props.goToDatabasePage()}
                    >
                        <DatabaseViewer
                            expand={true}
                            data={this.props.dbData}
                        />
                    </WidgetWrapper>
                </div>
            </div>
        );
    }
}

const mapState = state => ({
    dbData: state.dbData,
    dbStats: state.dbStatsData
});

const mapDispatch = dispatch =>
    bindActionCreators(
        {
            goToStatsPage: () => push('/stats'),
            goToDatabasePage: () => push('/database')
        },
        dispatch
    );

export default connect(mapState, mapDispatch)(DashboardContainer);
