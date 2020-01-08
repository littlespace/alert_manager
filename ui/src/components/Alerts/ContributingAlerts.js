import React from "react";
import styled from "styled-components";

import { capitalize, SEVERITY_COLORS, STATUS_COLOR } from "../../library/utils";
import { PRIMARY } from "../../styles/styles";

const COLUMNS = ["status", "name", "site", "device", "entity"];

const AlertGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(5, minmax(auto, 400px));
  grid-auto-rows: auto;
  align-items: center;
  grid-row-gap: 1em;
  padding: 0 1em 1em;
`;

const DetailContainer = styled.div`
  padding: ${({ status }) => (!status ? "1em 0" : null)};
  border-bottom: 1px solid ${PRIMARY};
  color: ${({ status, severity }) =>
    status
      ? null
      : SEVERITY_COLORS[severity.toLowerCase()]["background-color"]};
  cursor: pointer;
`;

const Status = styled.div`
  border-radius: 5px;
  background: ${({ status }) =>
    STATUS_COLOR[status.toLowerCase()]["background-color"]};
  color: ${({ status }) => STATUS_COLOR[status.toLowerCase()]["color"]};
  width: 50%;
  text-align: center;
  padding: 1em;
`;

function handleDetailOnClick(id) {
  window.open(`${window.location.origin}/alert/${id}`);
}

function Detail({ alert }) {
  return COLUMNS.map((col, idx) => {
    let detail;
    let status = false;
    if (col.toLowerCase() === "status") {
      status = true;
      detail = <Status status={alert.status}>{alert[col]}</Status>;
    } else {
      detail = alert[col];
    }

    return (
      <DetailContainer
        key={idx}
        onClick={() => handleDetailOnClick(alert.id)}
        severity={alert.severity}
        status={status}
      >
        {detail}
      </DetailContainer>
    );
  });
}

function ContributingAlerts({ alerts }) {
  return (
    <AlertGrid>
      {COLUMNS.map((col, idx) => (
        <h3 key={idx}>{capitalize(col)}</h3>
      ))}
      {alerts.map((alert, idx) => {
        return <Detail key={idx} alert={alert} />;
      })}
    </AlertGrid>
  );
}

export default ContributingAlerts;
