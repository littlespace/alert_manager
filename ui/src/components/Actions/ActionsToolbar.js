import React from "react";
import styled from "styled-components";

import { PRIMARY } from "../../styles/styles";

import ActionInput from "./ActionInput";
import TeamViewSelect from "./TeamViewSelect";
import TimeRangeSelect from "./TimeRangeSelect";

const Toolbar = styled.div`
  display: flex;
  background-color: ${PRIMARY};
  align-items: center;
  justify-content: flex-end;
  flex-wrap: wrap;
  padding: 1em 0.5em;

  & > select,
  input[type="text"] {
    margin-bottom: 0.5em;
  }
`;

export default function ActionsToolbar(props) {
  return (
    <Toolbar>
      <ActionInput
        filterType={"name"}
        placeholder={"Search by alert name..."}
        {...props}
      />
      <TimeRangeSelect {...props} />
      <TeamViewSelect {...props} />
    </Toolbar>
  );
}
