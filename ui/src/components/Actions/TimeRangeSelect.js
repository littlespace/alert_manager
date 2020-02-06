import React from "react";
import styled from "styled-components";
import { useLocation } from "react-router-dom";

import { getSearchOptionsByKey } from "../../library/utils";

import ActionSelect from "./ActionSelect";

const TIME_RANGES_H = [
  { label: "1H", value: "1h" },
  { label: "24H", value: "24h" },
  { label: "7 Days", value: "168h" },
  { label: "14 Days", value: "336h" },
  { label: "30 Days", value: "720h" }
];

const TimeRange = styled.div`
  flex: 0 0 200px;
  margin: auto 1em;
`;

function getValueFromOptions(options) {
  // No option is selected
  if (options.length === 0) {
    return null;
  }

  let label;
  let value = options[0];
  TIME_RANGES_H.forEach(timeRange => {
    if (timeRange.value === value) {
      label = timeRange.label;
    }
  });

  return { label: label, value: value };
}

function getSelectedOption(actionType, location) {
  let options = getSearchOptionsByKey(actionType, location);
  return getValueFromOptions(options);
}

export default function TimeRangeSelect({ setSearch }) {
  let location = useLocation();
  const actionType = "timerange";

  return (
    <TimeRange>
      <ActionSelect
        actionType={actionType}
        options={TIME_RANGES_H}
        placeholder={"Select Time Range"}
        searchOnChange={true}
        setSearch={setSearch}
        value={getSelectedOption(actionType, location)}
      />
    </TimeRange>
  );
}
