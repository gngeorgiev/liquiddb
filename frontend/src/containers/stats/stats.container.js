import React, { Component } from 'react';
import { connect } from 'react-redux';
import { Container, Row, Col } from 'react-grid-system';
import List, { ListItem, ListItemText } from 'material-ui/List';

import ConnectionsCount from '../../components/connections-count/connections-count.component';
import WidgetWrapper from '../../components/widget-wrapper/widget-wrapper.component';

class Stats extends Component {
    render() {
        const { dbStats } = this.props;
        return (
            <Container className="grid-container">
                <Row>
                    <Col md={4}>
                        <WidgetWrapper title="Connections count">
                            <ConnectionsCount
                                count={dbStats.connections.length}
                            />
                        </WidgetWrapper>
                    </Col>
                </Row>
                <Row>
                    <Col md={4}>
                        <WidgetWrapper title="Connections">
                            <div style={{ maxHeight: 200, overflow: 'auto' }}>
                                <List>
                                    {dbStats.connections.map(c =>
                                        <ListItem key={c}>
                                            <ListItemText primary={c} />
                                        </ListItem>
                                    )}
                                </List>
                            </div>
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
