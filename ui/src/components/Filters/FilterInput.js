import React, { useContext, useEffect, useState } from "react";
import styled from "styled-components";

import { FilterContext } from "../contexts/FilterContext";
import { PRIMARY, HIGHLIGHT } from "../../styles/styles";
import { TABLE_ACTIONS } from "../../library/utils";

const Input = styled.input`
  position: absolute;
  top: 17px;
  right: 230px;
  background-color: ${PRIMARY};
  color: ${HIGHLIGHT};
  border: 0px solid ${PRIMARY};
  border-bottom: 1px solid ${HIGHLIGHT};
  font-size: medium;
  ::placeholder {
    color: ${HIGHLIGHT};
  }
`;

const onChangeHandler = (filters, setFilters, value, setValue, type) => {
  setValue(value);
  setFilters(filters => ({ ...filters, [type]: value }));
};

function FilterInput({
  filterType,
  placeholder,
  tableMutationState,
  tableMutationDispatch
}) {
  const { filters, setFilters } = useContext(FilterContext);
  const [value, setValue] = useState("");

  useEffect(() => {
    if (tableMutationState.clearInput) {
      setValue("");
      delete filters[filterType];
      setFilters(filters => ({ ...filters }));
      tableMutationDispatch({ type: TABLE_ACTIONS.UNSET_CLEAR_INPUT });
    }
  }, [tableMutationState.clearInput]);

  return (
    <Input
      value={value}
      onChange={e =>
        onChangeHandler(
          filters,
          setFilters,
          e.target.value,
          setValue,
          filterType
        )
      }
      placeholder={placeholder}
    ></Input>
  );
}

export default FilterInput;
