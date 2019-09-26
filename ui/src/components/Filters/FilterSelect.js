import React from "react";
import styled from "styled-components";

import { PRIMARY, HIGHLIGHT } from "../../styles/styles";
import { TABLE_ACTIONS } from "../../library/utils";

const TIME_RANGES_H = [
  { label: "1H", value: "1h" },
  { label: "24H", value: "24h" },
  { label: "7 Days", value: "168h" },
  { label: "14 Days", value: "336h" },
  { label: "30 Days", value: "720h" }
];

const Select = styled.select`
  position: absolute;
  top: 17px;
  right: 50px;
  background-color: ${PRIMARY};
  color: ${HIGHLIGHT};
  border: 0px solid ${PRIMARY};
  border-bottom: 1px solid ${HIGHLIGHT};
  font-size: medium;
  ::placeholder {
    color: ${HIGHLIGHT};
  }
`;

function FilterSelect({ tableMutationDispatch }) {
  return (
    <Select
      defaultValue={"DEFAULT"}
      onChange={e =>
        tableMutationDispatch({
          type: TABLE_ACTIONS.SET_TIMERANGE,
          timeRange: e.target.value
        })
      }
    >
      <option value="DEFAULT" disabled hidden>
        Select Time Range
      </option>
      {TIME_RANGES_H.map((time, index) => (
        <option key={index} value={time["value"]}>
          {time["label"]}
        </option>
      ))}
    </Select>
  );
}

export default FilterSelect;
