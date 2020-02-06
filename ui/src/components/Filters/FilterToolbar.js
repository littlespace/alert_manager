import React from "react";
import styled from "styled-components";

import { PRIMARY, SECONDARY, HIGHLIGHT, ROBLOX } from "../../styles/styles";

import FilterMultiSelect from "./FilterMultiSelect";
import { IoMdSearch } from "react-icons/io";

// This name must match the column assesor field.
const MULTI_FILTERS = ["severity", "status", "device", "site", "source"];

const GridStyle = styled.div`
  display: grid;
  background-color: ${PRIMARY};
  grid-template-columns: repeat(6, 1fr);
  grid-gap: 1em;
  align-items: center;
  padding: 3em 0.5em;
`;

const Search = styled(IoMdSearch)`
  color: ${HIGHLIGHT};
  font-size: 52px;
  border-radius: 100%;
  padding: 0.15em;
  cursor: pointer;
  transition: 0.2s;

  :hover {
    background-color: ${SECONDARY};
    color: ${ROBLOX};
  }
`;

function FilterToolbar({ setSearch }) {
  return (
    <>
      <GridStyle>
        {MULTI_FILTERS.map((filterType, index) => {
          return (
            <FilterMultiSelect
              key={index}
              filterType={filterType}
              placeholder={filterType}
            />
          );
        })}
        <Search onClick={() => setSearch(true)} />
      </GridStyle>
    </>
  );
}

export default FilterToolbar;
