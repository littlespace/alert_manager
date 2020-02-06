import React, { useContext, useEffect } from "react";
import styled from "styled-components";
import { useLocation, useHistory } from "react-router-dom";

import { PRIMARY, HIGHLIGHT } from "../../styles/styles";
import { setSearchString, getSearchOptionsByKey } from "../../library/utils";

import { TableContext } from "../contexts/TableContext";

const Input = styled.input.attrs({ type: "text" })`
  margin: auto 0 auto 0;
  background-color: ${PRIMARY};
  color: ${HIGHLIGHT};
  border: 0px solid ${PRIMARY};
  border-bottom: 1px solid ${HIGHLIGHT};
  width: 350px;
  font-size: medium;
  ::placeholder {
    color: ${HIGHLIGHT};
  }
`;

function getSelectedOption(filterType, location) {
  let options = getSearchOptionsByKey(filterType, location);
  // No option is selected
  if (options.length === 0) {
    return null;
  }
  return options.length === 0 ? null : options[0];
}

function onChangeHandler(value, type, location, history, setFilter) {
  setSearchString(type, value, location, history);
  setFilter(type, value);
}

function ActionInput({ filterType, placeholder }) {
  const { setFilter } = useContext(TableContext);

  let history = useHistory();
  let location = useLocation();

  // On first mount, check the URL and see if we need to set the filter initially
  useEffect(() => {
    setFilter(filterType, getSelectedOption(filterType, location));
  }, []);

  return (
    <Input
      value={getSelectedOption(filterType, location)}
      onChange={e => {
        onChangeHandler(
          e.target.value,
          filterType,
          location,
          history,
          setFilter
        );
      }}
      placeholder={placeholder}
    ></Input>
  );
}

export default ActionInput;
