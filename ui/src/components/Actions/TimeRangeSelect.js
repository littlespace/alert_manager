import React from "react";

import { TABLE_ACTIONS } from "../../library/utils";

import ActionSelect from "./ActionSelect";

const TIME_RANGES_H = [
  { label: "1H", value: "1h" },
  { label: "24H", value: "24h" },
  { label: "7 Days", value: "168h" },
  { label: "14 Days", value: "336h" },
  { label: "30 Days", value: "720h" }
];

export default function TimeRangeSelect({ ...props }) {
  return (
    <ActionSelect actionType={TABLE_ACTIONS.SET_TIMERANGE} {...props}>
      <option value="DEFAULT" disabled hidden>
        Select Time Range
      </option>
      {TIME_RANGES_H.map((time, index) => (
        <option key={index} value={time["value"]}>
          {time["label"]}
        </option>
      ))}
    </ActionSelect>
  );
}
