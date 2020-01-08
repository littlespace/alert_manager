import React from "react";
import styled from "styled-components";

import { PRIMARY } from "../../styles/styles";

const HistoryGrid = styled.div`
  display: grid;
  grid-gap: 1em 0;
  grid-template-columns: 1fr 4fr;
  grid-auto-rows: auto;
  padding: 0 1em 1em;
  font-family: monospace;

  & > div:last-of-type {
    border: none;
  }

  & > div:nth-last-of-type(2) {
    border: none;
  }
`;

const Timestamp = styled.div`
  border-bottom: 2px solid ${PRIMARY};
  padding-bottom: 1em;
`;

const Event = styled.div`
  border-bottom: 2px solid ${PRIMARY};
  padding-bottom: 1em;
`;

const getTimestamp = timestamp => {
  const date = new Date(timestamp * 1000);
  return `${date.toLocaleDateString()} ${date.toLocaleTimeString()}`;
};

function AlertHistory({ history }) {
  return (
    <HistoryGrid>
      <h3>Time</h3>
      <h3>Event</h3>
      {history.map(({ event, timestamp }, idx) => (
        <React.Fragment key={idx}>
          <Timestamp>{getTimestamp(timestamp)}</Timestamp>
          <Event>{event}</Event>
        </React.Fragment>
      ))}
    </HistoryGrid>
  );
}

export default AlertHistory;
