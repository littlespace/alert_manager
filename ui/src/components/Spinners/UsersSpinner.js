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

export default function UsersSpinner() {
  let team = localStorage.getItem("user_team");
  return (
    <SpinnerContainer>
      <Spinner />
      <SpinnerTitle>{`Grabbing Users For ${team.toUpperCase()}`}</SpinnerTitle>
      <SpinnerComment>Let's Manage Those Users!</SpinnerComment>
    </SpinnerContainer>
  );
}
