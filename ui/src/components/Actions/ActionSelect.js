import React from "react";
import styled from "styled-components";

import { PRIMARY, HIGHLIGHT, SECONDARY } from "../../styles/styles";

const Select = styled.select`
  background-color: ${SECONDARY};
  color: ${HIGHLIGHT};
  border: 0px solid ${PRIMARY};
  border-bottom: 1px solid ${HIGHLIGHT};
  flex: 0 0 200px;
  margin: auto 0.5em;
  font-size: medium;
  padding: 0.2em;
  border-radius: 3px;
  ::placeholder {
    color: ${HIGHLIGHT};
  }
`;

export default function FilterSelect({
  actionType,
  tableMutationDispatch,
  children
}) {
  return (
    <Select
      defaultValue={"DEFAULT"}
      onChange={e =>
        tableMutationDispatch({
          type: actionType,
          value: e.target.value
        })
      }
    >
      {children}
    </Select>
  );
}
