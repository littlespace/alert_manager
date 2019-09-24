import React, { useContext } from "react";
import styled from "styled-components";

import { AlertManagerApi } from "../../library/AlertManagerApi";
import {
  CRITICAL,
  HIGHLIGHT,
  INFO,
  PRIMARY,
  SECONDARY
} from "../../styles/styles";
import { NotificationContext } from "../contexts/NotificationContext";
import { TableContext } from "../contexts/TableContext";

const ALERT_ACTION = {
  ACKNOWLEDGE: opts => api.alertAcknowledge(opts),
  CLEAR: opts => api.alertClear(opts)
};

const api = new AlertManagerApi();

const SelectedContainer = styled.div`
  display: flex;
  background-color: ${PRIMARY};
  position: sticky;
  z-index: 11;
  top: 0;
  padding: 0.75em;
  border-bottom: 1px groove ${HIGHLIGHT};
`;

const Selected = styled.div`
  font-size: small;
  padding: 0.5em;
  margin: auto 0.5em;
`;

const SelectedButton = styled.button`
  background-color: ${HIGHLIGHT};
  border: 1px solid ${HIGHLIGHT};
  color: ${SECONDARY};
  font-size: small;
  border-radius: 10px;
  padding: 0.5em 1em;
  margin: auto 0.5em;
  transition: 0.5s;
  :hover {
    background-color: ${SECONDARY};
    color: ${HIGHLIGHT};
    box-shadow: 1px 1px 3px ${HIGHLIGHT};
    cursor: pointer;
  }
`;

function handleOnClick(
  handlerFunc,
  rowSelectState,
  setNotificationBar,
  setNotificationColor,
  setNotificationMsg
) {
  let msg = null;
  let color = null;
  const user = api.getUsername();

  try {
    rowSelectState.ids.forEach(async id => {
      console.log(`${handlerFunc.name.toLowerCase()} Alert ID: ${id}`);
      await handlerFunc({ id: id, owner: user });
    });

    msg = `Successfully ${handlerFunc.name.toLowerCase()}'d ${
      rowSelectState.ids.length
    } alerts.`;
    color = INFO;
  } catch (error) {
    msg = `Failed to ${handlerFunc.name.toLowerCase()} alerts: ${error}`;
    color = CRITICAL;
    console.log(msg);
  }

  setNotificationMsg(msg);
  setNotificationColor(color);
  setNotificationBar(true);
}

export default function SelectToolbar() {
  const { rowSelectState } = useContext(TableContext);
  const {
    setNotificationMsg,
    setNotificationColor,
    setNotificationBar
  } = useContext(NotificationContext);

  return (
    <>
      <SelectedContainer>
        <Selected>Selected Rows: {rowSelectState.rows.length}</Selected>
        <SelectedButton
          onClick={() =>
            handleOnClick(
              ALERT_ACTION.ACKNOWLEDGE,
              rowSelectState,
              setNotificationBar,
              setNotificationColor,
              setNotificationMsg
            )
          }
        >
          Acknowledge
        </SelectedButton>
        <SelectedButton
          onClick={() =>
            handleOnClick(
              ALERT_ACTION.CLEAR,
              rowSelectState,
              setNotificationBar,
              setNotificationColor,
              setNotificationMsg
            )
          }
        >
          Clear
        </SelectedButton>
      </SelectedContainer>
    </>
  );
}
