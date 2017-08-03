import React, { Component } from 'react';
import { connect } from 'react-redux';
import { Container, Row, Col } from 'react-grid-system';

import ConnectionsCount from '../../components/connections-count/connections-count.component';
import WidgetWrapper from '../../components/widget-wrapper/widget-wrapper.component';

class Stats extends Component {
    render() {
        return (
            <Container className="grid-container">
                <Row>
                    <Col md={4}>
                        <WidgetWrapper title="Connections">
                            <ConnectionsCount
                                count={this.props.dbStats.connections.length}
                            />
                        </WidgetWrapper>
                    </Col>
                </Row>
            </Container>
        );
    }
}

const mapState = state => ({
    dbStats: state.dbStatsData
});

export default connect(mapState)(Stats);
