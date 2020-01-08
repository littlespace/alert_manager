import React from "react";
import styled from "styled-components";

import { HIGHLIGHT } from "../../styles/styles";

import Spinner from "./Spinner";

const SpinnerContainer = styled.div`
  display: flex;
  flex-direction: column;
  margin-top: 10px;
  height: 200px;
`;

const SpinnerTitle = styled.div`
  font-size: x-large;
  color: ${HIGHLIGHT};
  font-weight: 500;
  margin: 5px auto;
`;

const SpinnerComment = styled.div`
  font-size: medium;
  color: ${HIGHLIGHT};
  margin: 5px auto;
`;

function AlertSpinner() {
  return (
    <SpinnerContainer>
      <Spinner />
      <SpinnerTitle>Grabbing Alert Details</SpinnerTitle>
      <SpinnerComment>Are you excited because I am!</SpinnerComment>
    </SpinnerContainer>
  );
}

export default AlertSpinner;
