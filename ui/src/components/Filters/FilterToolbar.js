import React from "react";
import styled from "styled-components";

import { getAlertFilterOptions } from "../../library/utils";
import { PRIMARY } from "../../styles/styles";
import FilterMultiSelect from "./FilterMultiSelect";
import ActionsToolbar from "../Actions/ActionsToolbar";

// This name must match the column assesor field.
const MULTI_FILTERS = [
  "severity",
  "status",
  "device",
  "site",
  "source",
  "entity"
];

const GridStyle = styled.div`
  display: grid;
  background-color: ${PRIMARY};
  grid-template-columns: repeat(6, 1fr);
  grid-gap: 10px;
  padding: 3em 0.5em;
`;

function FilterToolbar({ alerts, ...props }) {
  return (
    <>
      <ActionsToolbar {...props} />
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
    </>
  );
}

export default FilterToolbar;
