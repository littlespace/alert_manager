import React, { useState, useEffect, useContext } from "react";
import styled from "styled-components";
import { withRouter, useLocation, useHistory } from "react-router-dom";

import { AlertManagerApi } from "../library/AlertManagerApi";
import { HIGHLIGHT } from "../styles/styles";
import {
  setSearchString,
  getSearchOptions,
  getSearchString,
  secondsToHms
} from "../library/utils";

import { NotificationContext } from "../components/contexts/NotificationContext";
import { TableProvider } from "../components/contexts/TableContext";

import AlertsSpinner from "../components/Spinners/AlertsSpinner";
import AlertsTable from "../components/AlertsTable/AlertsTable";
import COLUMNS from "../components/AlertsTable/columns";
import FilterToolbar from "../components/Filters/FilterToolbar";
import NotificationBar from "../components/Notifications/NotificationBar";
import ActionsToolbar from "../components/Actions/ActionsToolbar";

const api = new AlertManagerApi();

const Wrapper = styled.div`
  color: ${HIGHLIGHT};
`;

function convertTimestamps(data) {
  data.forEach(element => {
    // in unix time and javascript needs to read it in ms not seconds
    let start = new Date(element.start_time * 1000);
    element.start_time = start.toLocaleTimeString();
    element.start_date = start.toLocaleDateString();
    element.last_active = secondsToHms(element.last_active);
  });
}

function sortAlerts(alerts) {
  const severity = {
    CRITICAL: 3,
    WARN: 2,
    INFO: 1
  };

  // Sort based on last_active first
  alerts.sort((first, second) =>
    first.last_active < second.last_active
      ? 1
      : first.last_active > second.last_active
      ? -1
      : 0
  );

  // Then sort by severity
  alerts.sort((first, second) =>
    severity[first.severity] < severity[second.severity]
      ? 1
      : severity[first.severity] > severity[second.severity]
      ? -1
      : 0
  );
}

function getDefaultTeam(location, history) {
  const team = api.getTeam();
  setSearchString("team", team, location, history);

  return team;
}

// TODO: Add propTypes
function AlertsView(props) {
  const [loading, setLoading] = useState(false);
  const [alerts, setAlerts] = useState([]);
  const [teamList, setTeamList] = useState([]);
  const [search, setSearch] = useState(false);

  let location = useLocation();
  let history = useHistory();

  const {
    notificationBar,
    setNotificationBar,
    notificationColor,
    notificationMsg
  } = useContext(NotificationContext);

  const fetchAlerts = async () => {
    setLoading(true);
    let options = getSearchOptions(location);
    const results = await api.getAlertsList({
      limit: 5000,
      history: true,
      timerange_h: options.timerange || null,
      teams: options.team || [getDefaultTeam(location, history)],
      severity: options.severity || [],
      status: options.status || [],
      devices: options.device || [],
      sites: options.site || [],
      sources: options.source || []
    });

    // Default sorting
    sortAlerts(results);
    // Normalize the alerts for display in the UI
    convertTimestamps(results);

    setAlerts(results);
    setLoading(false);
    setSearch(false);
  };

  const fetchTeamList = async () => {
    const teams = await api.getTeamList();
    setTeamList(teams.map(team => team.Name));
  };

  // This get's all our data on the first mount
  useEffect(() => {
    fetchAlerts();
    fetchTeamList();
  }, []);

  // Trigger alert query when search is set to true
  useEffect(() => {
    if (search) {
      fetchAlerts();
    }
  }, [search]);

  // Used for our notifications bar
  useEffect(() => {
    if (notificationBar === true) {
      setTimeout(() => setNotificationBar(false), 10000);
      // We need to add a timeout to wait before fetching the data again. It takes
      // the backend a few seconds to clear/ack the alerts.
      setLoading(true);
      setTimeout(() => fetchAlerts(), 2000);
    }
  }, [notificationBar]);

  return (
    <>
      {/* <div style={{ color: "black" }}>
        <pre>
          <code>{console.log("Rendering the whole shabang")}</code>
        </pre>
      </div> */}
      <TableProvider data={alerts} columns={COLUMNS}>
        {notificationBar ? (
          <NotificationBar color={notificationColor} msg={notificationMsg} />
        ) : null}
        <Wrapper>
          <ActionsToolbar teamList={teamList} setSearch={setSearch} />
          <FilterToolbar setSearch={setSearch} />
          {loading ? <AlertsSpinner /> : <AlertsTable />}
        </Wrapper>
      </TableProvider>
    </>
  );
}

export default withRouter(AlertsView);
