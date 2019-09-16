import React, { useState, useEffect } from "react";
import { withRouter } from "react-router-dom";
import styled from "styled-components";

import { ALERT_STATUS } from "../static";
import { AlertManagerApi } from "../library/AlertManagerApi";
import { HIGHLIGHT } from "../styles/styles";
import AlertsTable from "../components/AlertsTable/AlertsTable";
import AlertsSpinner from "../components/Spinners/AlertsSpinner";
import FilterToolbar from "../components/Filters/FilterToolbar";
import { secondsToHms } from '../library/utils';

const api = new AlertManagerApi();

// TODO: Add filter options in the Columns Def so we can dynamically get the
// list instead of hardcoded in "AlertsView" obj
const COLUMNS = [
  { accessor: "id", Header: "Id", show: false },
  { accessor: "history", Header: "History", show: false },
  { accessor: "severity", Header: "Severity", filter: SelectMultiColumnFilter },
  { accessor: "status", Header: "Status", filter: SelectMultiColumnFilter },
  { accessor: "name", Header: "Name" },
  { accessor: "site", Header: "Site", filter: SelectMultiColumnFilter },
  { accessor: "device", Header: "Device", filter: SelectMultiColumnFilter },
  { accessor: "entity", Header: "Entity", filter: SelectMultiColumnFilter },
  { accessor: "source", Header: "Source", filter: SelectMultiColumnFilter },
  { accessor: "start_date", Header: "Start Date", show: true },
  { accessor: "start_time", Header: "Start Time", show: true },
  { accessor: "last_active", Header: "Last Active" },
  { accessor: "details", Header: "Details" }
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
    element.last_active = secondsToHms(element.last_active);
  });
}

function addDetails(data) {
  // Add text string to details column
  data.forEach(element => {
    element.details = "More Info";
  });
}


// TODO: Add propTypes
function AlertsView(props) {
  const [loading, setLoading] = useState(false);
  const [alerts, setAlerts] = useState([]);
  const [filters, setFilters] = useState({});
  const [timeRange, setTimeRange] = useState(0);
  const [status, setStatus] = useState([
    ALERT_STATUS["active"],
    ALERT_STATUS["suppressed"]
  ]);

  useEffect(() => {
    const fetchAlerts = async () => {
      setLoading(true);
      const results = await api.getAlertsList({
        limit: 5000,
        status: status,
        timerange_h: timeRange,
        history: true,
      });

      console.log(results)
      // Normalize the alerts for display in the UI
      convertTimestamps(results);
      addDetails(results);

      setAlerts(results);
      setLoading(false);
    };

    fetchAlerts();
  }, [status, timeRange]);

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
          setStatus={setStatus}
        />
        {loading ? <AlertsSpinner /> : ( 
          <AlertsTable
            filters={filters}
            columns={COLUMNS}
            data={alerts}
            history={props.history}
          />
        )}
      </Wrapper>
    </>
  );
}

export default withRouter(AlertsView);
