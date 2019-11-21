import React from "react";

import { TABLE_ACTIONS, capitalize } from "../../library/utils";

import FilterSelect from "./ActionSelect";

export default function TimeRangeSelect({ teamList, ...props }) {
  return (
    <FilterSelect actionType={TABLE_ACTIONS.SET_TEAM} {...props}>
      <option value="DEFAULT" disabled hidden>
        Select Team
      </option>
      <option value="">All</option>
      {teamList.map((team, index) => (
        <option key={index} value={team}>
          {capitalize(team)}
        </option>
      ))}
    </FilterSelect>
  );
}
