import React, { useState, useEffect, useContext, useReducer } from "react";
import styled from "styled-components";
import { withRouter } from "react-router-dom";

import { ALERT_STATUS } from "../static";
import { AlertManagerApi } from "../library/AlertManagerApi";
import { FilterProvider } from "../components/contexts/FilterContext";
import { HIGHLIGHT } from "../styles/styles";
import { NotificationContext } from "../components/contexts/NotificationContext";
import { secondsToHms } from "../library/utils";
import { TABLE_ACTIONS } from "../library/utils";
import { TableProvider } from "../components/contexts/TableContext";
import AlertsSpinner from "../components/Spinners/AlertsSpinner";
import AlertsTable from "../components/AlertsTable/AlertsTable";
import COLUMNS from "../components/AlertsTable/columns";
import FilterToolbar from "../components/Filters/FilterToolbar";
import NotificationBar from "../components/NotificationBar";

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

function tableMutationReducer(state, action) {
  switch (action.type) {
    case TABLE_ACTIONS.SET_CLEAR_MUTATIONS:
      return {
        ...state,
        clearMultiselect: true,
        clearInput: true,
        clearSelection: true
      };
    case TABLE_ACTIONS.UNSET_CLEAR_MUTATIONS:
      return {
        ...state,
        clearMultiselect: false,
        clearInput: false,
        clearSelection: false
      };
    case TABLE_ACTIONS.SET_CLEAR_MULTISELECT:
      return { ...state, clearMultiselect: true };
    case TABLE_ACTIONS.UNSET_CLEAR_MULTISELECT:
      return { ...state, clearMultiselect: false };
    case TABLE_ACTIONS.SET_CLEAR_INPUT:
      return { ...state, clearInput: true };
    case TABLE_ACTIONS.UNSET_CLEAR_INPUT:
      return { ...state, clearInput: false };
    case TABLE_ACTIONS.SET_CLEAR_SELECTION:
      return { ...state, clearSelection: true };
    case TABLE_ACTIONS.UNSET_CLEAR_SELECTION:
      return { ...state, clearSelection: false };
    case TABLE_ACTIONS.SET_TIMERANGE:
      return { ...state, timeRange: action.timeRange };
    case TABLE_ACTIONS.SET_STATUS:
      return { ...state, status: action.status };
  }
}

// TODO: Add propTypes
function AlertsView(props) {
  const [loading, setLoading] = useState(false);
  const [alerts, setAlerts] = useState([]);
  const {
    notificationBar,
    setNotificationBar,
    notificationColor,
    notificationMsg
  } = useContext(NotificationContext);
  const [tableMutationState, tableMutationDispatch] = useReducer(
    tableMutationReducer,
    {
      clearMultiselect: false,
      clearInput: false,
      clearSelection: false,
      timeRange: 0,
      status: [ALERT_STATUS["active"], ALERT_STATUS["suppressed"]]
    }
  );

  const fetchAlerts = async () => {
    setLoading(true);
    const results = await api.getAlertsList({
      limit: 5000,
      status: tableMutationState.status,
      timerange_h: tableMutationState.timeRange,
      history: true
    });

    // Normalize the alerts for display in the UI
    convertTimestamps(results);

    setAlerts(results);
    setLoading(false);
  };

  useEffect(() => {
    tableMutationDispatch({ type: TABLE_ACTIONS.SET_CLEAR_MUTATIONS });
    fetchAlerts();
  }, [tableMutationState.status, tableMutationState.timeRange]);

  useEffect(() => {
    if (notificationBar === true) {
      tableMutationDispatch({ type: TABLE_ACTIONS.SET_CLEAR_MUTATIONS });
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
      <FilterProvider>
        <TableProvider data={alerts} columns={COLUMNS}>
          {notificationBar ? (
            <NotificationBar color={notificationColor} msg={notificationMsg} />
          ) : null}
          <Wrapper>
            <FilterToolbar
              alerts={alerts}
              tableMutationDispatch={tableMutationDispatch}
              tableMutationState={tableMutationState}
            />
            {loading ? (
              <AlertsSpinner />
            ) : (
              <AlertsTable
                tableMutationDispatch={tableMutationDispatch}
                tableMutationState={tableMutationState}
              />
            )}
          </Wrapper>
        </TableProvider>
      </FilterProvider>
    </>
  );
}

export default withRouter(AlertsView);
