import React from 'react';
import { Card } from 'material-ui/Card';
import { ViewTitle } from 'admin-on-rest/lib/mui';
import JSONTree from 'react-json-tree';

import './dashboard.index.css';

export const Dashboard = () =>
    <Card>
        <ViewTitle title="Explorer" />
        <JSONTree
            theme="solarized"
            data={{
                test: 1
            }}
        />
    </Card>;
