import React from 'react';
import styled from "styled-components";

import { SECONDARY, PRIMARY } from "../../styles/styles";
import EllipsisSpinner from "./EllipsisSpinner";

const SpinnerContainer = styled.div`
  position: relative;
`;

const SpinnerTitle = styled.div`
  position: absolute;
  font-size: x-large;
  text-align: center;
  color: ${PRIMARY};
  font-weight: 500;
  top: 60px;
  left: 45%;
`;

const SpinnerComment = styled.div`
  position: absolute;
  font-size: medium;
  color: ${SECONDARY};
  top: 95px;
  left: 40%;
`;

function AlertsSpinner() {
  return (
    <SpinnerContainer>
      <EllipsisSpinner color={SECONDARY} />
      <SpinnerTitle>Loading Alerts</SpinnerTitle>
      <SpinnerComment>
        Grab some coffee. It's gonna be a long night.
      </SpinnerComment>
    </SpinnerContainer>
  );
}

export default AlertsSpinner;