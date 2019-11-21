import React, { useState } from "react";
import styled from "styled-components";

import { ALERT_STATUS } from "../../static";
import { PRIMARY } from "../../styles/styles";
import { TABLE_ACTIONS } from "../../library/utils";

import ActionInput from "./ActionInput";
import Historical from "./Historical";
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

function handleCheckboxClick(event, setChecked, tableMutationDispatch) {
  setChecked(event.target.checked);
  // Add "historical" alerts when checked.
  if (event.target.checked === true) {
    tableMutationDispatch({
      type: TABLE_ACTIONS.SET_STATUS,
      status: [
        ALERT_STATUS["active"],
        ALERT_STATUS["suppressed"],
        ALERT_STATUS["cleared"],
        ALERT_STATUS["expired"]
      ]
    });
  } else {
    tableMutationDispatch({
      type: TABLE_ACTIONS.SET_STATUS,
      status: [ALERT_STATUS["active"], ALERT_STATUS["suppressed"]]
    });
  }
}

export default function ActionsToolbar({ ...props }) {
  const [checked, setChecked] = useState(false);

  return (
    <Toolbar>
      <Historical
        onChangeHandler={e =>
          handleCheckboxClick(e, setChecked, props.tableMutationDispatch)
        }
        checked={checked}
        title={"Historical"}
      />
      <TimeRangeSelect {...props} />
      <TeamViewSelect {...props} />
      <ActionInput
        filterType={"name"}
        placeholder={"Search by alert name..."}
        {...props}
      />
    </Toolbar>
  );
}
