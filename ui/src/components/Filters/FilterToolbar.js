import React, { useState } from "react";
import styled from "styled-components";

import PageviewRoundedIcon from "@material-ui/icons/PageviewRounded";

import { ALERT_STATUS } from "../../static";
import { getAlertFilterOptions, TABLE_ACTIONS } from "../../library/utils";
import { PRIMARY } from "../../styles/styles";
import FilterCheckbox from "./FilterCheckbox";
import FilterInput from "./FilterInput";
import FilterMultiSelect from "./FilterMultiSelect";
import FilterSelect from "./FilterSelect";

// This name must match the column assesor field.
const MULTI_FILTERS = [
  "severity",
  "status",
  "device",
  "site",
  "source",
  "entity"
];

const Toolbar = styled.div`
  background-color: ${PRIMARY};
  padding: 1.5em 2em 0.5em;
  position: relative;
`;

const Icon = styled.span`
  display: inline-flex;
  position: absolute;
  top: 17px;
  right: 402px;
  vertical-align: middle;
`;

const GridStyle = styled.div`
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  grid-gap: 10px;
  padding-top: 45px;
  padding-bottom: 30px;
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

function FilterToolbar({ alerts, ...props }) {
  const [checked, setChecked] = useState(false);

  return (
    <Toolbar>
      <FilterCheckbox
        onChangeHandler={e =>
          handleCheckboxClick(e, setChecked, props.tableMutationDispatch)
        }
        checked={checked}
        title={"Historical"}
      />
      <Icon>
        <PageviewRoundedIcon />
      </Icon>
      <FilterInput
        filterType={"name"}
        placeholder={"Search by alert name..."}
        {...props}
      />
      <FilterSelect {...props} />
      <GridStyle>
        {MULTI_FILTERS.map((filterType, index) => {
          return (
            <FilterMultiSelect
              key={index}
              filterType={filterType}
              options={getAlertFilterOptions(alerts, filterType)}
              placeholder={filterType}
              {...props}
            />
          );
        })}
      </GridStyle>
    </Toolbar>
  );
}

export default FilterToolbar;
