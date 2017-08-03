import React, { Component } from 'react';
import { connect } from 'react-redux';
import { push } from 'react-router-redux';
import { bindActionCreators } from 'redux';
import { Container, Row, Col } from 'react-grid-system';

import DatabaseViewer from '../../components/database-viewer/database-viewer.component';
import ConnectionsCount from '../../components/connections-count/connections-count.component';
import WidgetWrapper from '../../components/widget-wrapper/widget-wrapper.component';

import './dashboard.container.css';

class DashboardContainer extends Component {
    render() {
        return (
            <Container className="grid-container">
                <Row>
                    <Col md={4}>
                        <WidgetWrapper
                            title="Connections"
                            routeName="Stats"
                            actionPressed={() => this.props.goToStatsPage()}
                        >
                            <ConnectionsCount
                                count={this.props.dbStats.connections.length}
                            />
                        </WidgetWrapper>
                    </Col>
                </Row>
                <Row>
                    <Col>
                        <WidgetWrapper
                            title="Database View"
                            routeName="Database"
                            actionPressed={() => this.props.goToDatabasePage()}
                        >
                            <DatabaseViewer
                                expand={true}
                                data={this.props.dbData}
                            />
                        </WidgetWrapper>
                    </Col>
                </Row>
            </Container>
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
