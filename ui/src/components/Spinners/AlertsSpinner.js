import React from 'react';
import styled from "styled-components";

import { SECONDARY, PRIMARY } from "../../styles/styles";
import Spinner from "./Spinner";

const SpinnerContainer = styled.div`
  display: flex;
  flex-direction: column;
  margin-top: 10px;
  height: 200px;
`;

const SpinnerTitle = styled.div`
  font-size: x-large;
  color: ${PRIMARY};
  font-weight: 500;
  margin: 5px auto;
`;

const SpinnerComment = styled.div`
  font-size: medium;
  color: ${SECONDARY};
  margin: 5px auto;
`;

function AlertsSpinner() {
  return (
    <SpinnerContainer>
      <Spinner />
      <SpinnerTitle>Loading Alerts</SpinnerTitle>
      <SpinnerComment>
        Grab some coffee. It's gonna be a long night.
      </SpinnerComment>
    </SpinnerContainer>
  );
}

export default AlertsSpinner;