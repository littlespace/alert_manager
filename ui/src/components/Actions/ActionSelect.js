import React from "react";
import ReactSelect from "react-select";
import { useHistory, useLocation } from "react-router-dom";

import { PRIMARY, HIGHLIGHT, SECONDARY } from "../../styles/styles";
import { setSearchString } from "../../library/utils";

const ReactSelectStyles = {
  control: styles => ({ ...styles, backgroundColor: SECONDARY }),
  option: (styles, state) => ({
    ...styles,
    color: state.isFocused ? HIGHLIGHT : PRIMARY,
    backgroundColor: state.isFocused ? PRIMARY : HIGHLIGHT
  }),
  menu: styles => ({ ...styles, zIndex: 20 }),
  placeholder: styles => ({ ...styles, color: HIGHLIGHT, fontSize: "0.875em" }),
  input: styles => ({ ...styles, color: HIGHLIGHT }),
  dropdownIndictor: styles => ({ ...styles, color: HIGHLIGHT }),
  singleValue: styles => ({ ...styles, color: HIGHLIGHT })
};

function onChangeHandler(
  value,
  type,
  location,
  history,
  setSearch,
  searchOnChange
) {
  value = value ? value.value : null;
  setSearchString(type, value, location, history);
  setSearch(searchOnChange);
}

export default function ActionSelect({
  actionType,
  options,
  placeholder,
  searchOnChange,
  setSearch,
  value
}) {
  let history = useHistory();
  let location = useLocation();

  return (
    <ReactSelect
      isClearable
      isSearchable
      placeholder={placeholder}
      options={options}
      onChange={value =>
        onChangeHandler(
          value,
          actionType,
          location,
          history,
          setSearch,
          searchOnChange
        )
      }
      value={value}
      styles={ReactSelectStyles}
    />
  );
}
