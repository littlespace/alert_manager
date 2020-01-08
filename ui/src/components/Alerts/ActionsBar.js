import React, { useState, useContext } from "react";
import styled, { css } from "styled-components";

import { NotificationContext } from "../contexts/NotificationContext";

import {
  CRITICAL,
  HIGHLIGHT,
  INFO,
  PRIMARY,
  SECONDARY
} from "../../styles/styles";
import {
  SEVERITY_COLORS,
  SEVERITY_LEVELS,
  STATUS_COLOR
} from "../../library/utils";

import { FaRegCheckCircle, FaRegDotCircle, FaRegClock } from "react-icons/fa";
import ReactTooltip from "react-tooltip";
import Select from "react-select";

const Wrapper = styled.div`
  display: flex;
  align-items: center;
  padding: 1em;
  margin-left: 1em;
  justify-content: flex-start;
`;

const Status = styled.div`
  background: ${({ status }) =>
    STATUS_COLOR[status.toLowerCase()]["background-color"]};
  color: ${({ status }) => STATUS_COLOR[status.toLowerCase()]["color"]};
  padding: 1em;
  border-radius: 5px;
`;

const SeverityColor = css`
  color: ${({ value }) =>
    SEVERITY_COLORS[value.value.toLowerCase()]["background-color"]};
`;

const Severity = styled(Select)`
  margin: auto auto auto 1em;
  color: ${PRIMARY};
  cursor: pointer;
  width: 9em;
  border-radius: 5px;

  .react-select__single-value {
    color: unset;
  }

  .react-select__control {
    border-radius: 5px;
    border: 3px solid ${PRIMARY};
    background-color: ${({ value }) =>
      SEVERITY_COLORS[value.value.toLowerCase()]["background-color"]};
    min-height: unset;
    font-size: 1em;
    padding: calc(0.5em - 3px);
    transition: 0.2s ease-in-out;

    :hover {
      background-color: ${PRIMARY};
      border: 3px solid ${PRIMARY};
      ${SeverityColor}
    }
    :hover .react-select__indicator {
      ${SeverityColor}
    }

    :hover .react-select__indicator-separator {
      background-color: ${({ value }) =>
        SEVERITY_COLORS[value.value.toLowerCase()]["background-color"]};
    }
  }

  .react-select__control--menu-is-open {
    background-color: ${PRIMARY};
    border: 3px solid ${PRIMARY};
    ${SeverityColor}

    .react-select__indicator {
      ${SeverityColor}
    }

    .react-select__indicator-separator {
      background-color: ${({ value }) =>
        SEVERITY_COLORS[value.value.toLowerCase()]["background-color"]};
    }
  }

  .react-select__indicator {
    color: ${PRIMARY};
  }

  .react-select__indicator-separator {
    color: ${({ value }) =>
      SEVERITY_COLORS[value.value.toLowerCase()]["background-color"]};
    background-color: ${PRIMARY};
  }

  .react-select__menu {
    background-color: ${PRIMARY};
    ${SeverityColor}
  }

  /* need to add "div" to make this more specific  than '__option' */
  div .react-select__option--is-focused {
    color: ${HIGHLIGHT};
    background-color: ${SECONDARY};
  }

  .react-select__option {
    background-color: ${PRIMARY};
  }
`;

const ActionStyle = css`
  font-size: 3.3em;
  margin: 0 0.2em;
  color: ${PRIMARY};
  padding: 0.1em;
  cursor: pointer;

  :hover {
    transition: 0.3s;
    border-radius: 100%;
    background-color: ${PRIMARY};
    color: ${({ severity }) =>
      SEVERITY_COLORS[severity.toLowerCase()]["background-color"]};
  }
`;

const Acknowledge = styled(FaRegCheckCircle)`
  ${ActionStyle}
`;

const Clear = styled(FaRegDotCircle)`
  ${ActionStyle}
`;

const Suppress = styled(FaRegClock)`
  ${ActionStyle}
`;

const Tooltip = styled(ReactTooltip)`
  &.type-dark.place-bottom {
    color: ${HIGHLIGHT};
    background-color: ${PRIMARY};
    padding: 1em;
  }
`;

function ActionsBar({ alert, api, setUpdated, setShowSuppressForm }) {
  const [severity, setSeverity] = useState(alert.severity);

  const {
    setNotificationColor,
    setNotificationBar,
    setNotificationMsg
  } = useContext(NotificationContext);

  const updateSeverity = async value => {
    setSeverity(value);
    try {
      await api.updateAlertSeverity({
        id: alert.id,
        severity: value
      });
      setNotificationMsg(
        `Severity was updated to ${value} for alert id: ${alert.id}`
      );
      setNotificationColor(
        SEVERITY_COLORS[value.toLowerCase()]["background-color"]
      );
    } catch (err) {
      setNotificationColor(CRITICAL);
      setNotificationMsg(String(err));
    }

    setNotificationBar(true);
    setUpdated(true);
  };

  const clearAlert = async () => {
    try {
      await api.alertClear({ id: alert.id });
      setNotificationMsg(`Alert id: ${alert.id} was successfully cleared.`);
      setNotificationColor(INFO);
    } catch (err) {
      setNotificationColor(CRITICAL);
      setNotificationMsg(String(err));
    }

    setNotificationBar(true);
    setUpdated(true);
  };

  const acknowledgeAlert = async () => {
    try {
      await api.alertAcknowledge({ id: alert.id });
      setNotificationMsg(
        `Alert id: ${alert.id} was successfully acknowledged.`
      );
      setNotificationColor(INFO);
    } catch (err) {
      setNotificationColor(CRITICAL);
      setNotificationMsg(String(err));
    }

    setNotificationBar(true);
    setUpdated(true);
  };

  return (
    <Wrapper>
      <Status status={alert.status}>{alert.status}</Status>
      <Severity
        classNamePrefix={"react-select"}
        value={{ label: severity, value: severity }}
        options={SEVERITY_LEVELS.map(sev => ({ label: sev, value: sev }))}
        autoSize={true}
        onChange={({ value }) => updateSeverity(value)}
      />
      <Acknowledge
        data-tip={"Acknowledge"}
        severity={alert.severity}
        onClick={() => acknowledgeAlert()}
      />
      <Clear
        data-tip={"Clear"}
        severity={alert.severity}
        onClick={() => clearAlert()}
      />
      <Suppress
        data-tip={"Suppress"}
        severity={alert.severity}
        onClick={() => setShowSuppressForm(true)}
      />
      <Tooltip className={"react-tooltip"} place={"bottom"} effect={"solid"} />
    </Wrapper>
  );
}

export default ActionsBar;
