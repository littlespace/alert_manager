import React, { useState, useEffect } from 'react';
import { withRouter } from 'react-router-dom';
import styled from 'styled-components';

import { AlertManagerApi } from '../library/AlertManagerApi';
import { HIGHLIGHT } from '../styles/styles';
import AlertsTable from '../components/AlertsTable/AlertsTable';
import FilterToolbar from '../components/Filters/FilterToolbar';

const api = new AlertManagerApi();

// TODO: Add filter options in the Columns Def so we can dynamically get the 
// list instead of hardcoded in "AlertsView" obj
const COLUMNS = [
  { accessor: 'id', Header: 'Id', show: false },
  { accessor: 'severity', Header: 'Severity', filter: SelectMultiColumnFilter },
  { accessor: 'status', Header: 'Status', filter: SelectMultiColumnFilter },
  { accessor: 'name', Header: 'Name' },
  { accessor: 'site', Header: 'Site', filter: SelectMultiColumnFilter },
  { accessor: 'device', Header: 'Device', filter: SelectMultiColumnFilter },
  { accessor: 'entity', Header: 'Entity', filter: SelectMultiColumnFilter },
  { accessor: 'source', Header: 'Source', filter: SelectMultiColumnFilter },
  { accessor: 'start_date', Header: 'Start Date', show: true },
  { accessor: 'start_time', Header: 'Start Time', show: true },
  { accessor: 'last_active', Header: 'Last Active' },
  { accessor: 'details', Header: 'Details' },
];

const Wrapper = styled.div`
  color: ${HIGHLIGHT};
`;

function SelectMultiColumnFilter(rows, id, filterValue) {
  let filteredRows = rows.filter(row => {
    return filterValue.includes(row.values[id]);
  });
  return filteredRows;
}

function convertTimestamps(data) {
  data.forEach(element => {
    // in unix time and javascript needs to read it in ms not seconds
    let start = new Date(element.start_time * 1000);
    element.start_time = start.toLocaleTimeString();
    element.start_date = start.toLocaleDateString();
    element.last_active = `${start.toLocaleDateString()} ${start.toLocaleTimeString()}`;
  });
  return data;
}

function addDetails(data) {
  // Add text string to details column
  data.forEach(element => {
    element.details = 'More Info';
  });
  return data;
}

// TODO: Add propTypes
function AlertsView(props) {
  const [alerts, setAlerts] = useState([]);
  const [filters, setFilters] = useState({});
  const [timeRange, setTimeRange] = useState(0);

  useEffect(() => {
    api
      .getAlertsList({ limit: 2000, status: [], timerange_h: timeRange })
      .then(ret => convertTimestamps(ret))
      .then(ret => addDetails(ret))
      .then(ret => setAlerts(ret))
      .then(console.log('fetched alerts' + timeRange));
  }, [timeRange]);

  return (
    <>
      {/* <div>
        <pre>
          <code>{JSON.stringify(filters, null, 2)}</code>
        </pre>
      </div> */}
      <Wrapper>
        <FilterToolbar
          alerts={alerts}
          filters={filters}
          setFilters={setFilters}
          timeRange={timeRange}
          setTimeRange={setTimeRange}
        />
        <AlertsTable
          filters={filters}
          columns={COLUMNS}
          data={alerts}
          history={props.history}
        />
        }
      </Wrapper>
    </>
  );
}

export default withRouter(AlertsView);
