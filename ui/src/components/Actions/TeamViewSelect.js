import React from "react";
import styled from "styled-components";
import { useLocation } from "react-router-dom";

import { capitalize, getSearchOptionsByKey } from "../../library/utils";

import ActionSelect from "./ActionSelect";

const TeamView = styled.div`
  flex: 0 0 200px;
`;

function getSelectedOption(actionType, location) {
  let options = getSearchOptionsByKey(actionType, location);
  // No option is selected
  if (options.length === 0) {
    return null;
  }
  return { label: capitalize(options[0]), value: options[0] };
}

export default function TimeRangeSelect({ teamList, setSearch }) {
  let location = useLocation();
  const actionType = "team";

  return (
    <TeamView>
      <ActionSelect
        actionType={actionType}
        options={teamList.map(team => ({
          label: capitalize(team),
          value: team
        }))}
        placeholder={"Select Team"}
        searchOnChange={true}
        setSearch={setSearch}
        value={getSelectedOption(actionType, location)}
      />
    </TeamView>
  );
}
