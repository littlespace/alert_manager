import React, { useState, useEffect, useContext } from "react";
import ReactSelect from "react-select";

import {
  CRITICAL,
  HIGHLIGHT,
  INFO,
  PRIMARY,
  SECONDARY,
  WARN
} from "../../styles/styles";
import { FilterContext } from "../contexts/FilterContext";
import { TABLE_ACTIONS } from "../../library/utils";

// Using ReactSelects custom way to style
// TODO: Any way we can figure out how to do styled components here?
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
  multiValue: styles => {
    return {
      ...styles,
      backgroundColor: PRIMARY,
      color: HIGHLIGHT,
      fontSize: "0.875em"
    };
  },
  multiValueLabel: (styles, { data }) => {
    let color = HIGHLIGHT;
    if (data.value.toLowerCase() === "info") {
      color = INFO;
    } else if (data.value.toLowerCase() === "warn") {
      color = WARN;
    } else if (data.value.toLowerCase() === "critical") {
      color = CRITICAL;
    }
    return { ...styles, color: color };
  }
};

const onChangeHandler = (
  setSelectedOptions,
  filters,
  setFilters,
  filterType,
  values
) => {
  // e.g if user deleted all the options, we will have no entries
  if (values === null) {
    delete filters[filterType];
    setFilters(filters => ({ ...filters }));
    setSelectedOptions([]);
  } else {
    setSelectedOptions(values);
    const filterValues = [];
    values.forEach(e => filterValues.push(e.value));
    setFilters(filters => ({ ...filters, [filterType]: filterValues }));
  }
};

function FilterMultiSelect({
  filterType,
  options,
  placeholder,
  tableMutationState,
  tableMutationDispatch,
  ...props
}) {
  // [{ label: "ACTIVE", value: "ACTIVE"}]
  const { filters, setFilters } = useContext(FilterContext);
  const [selectedOptions, setSelectedOptions] = useState([]);

  useEffect(() => {
    if (tableMutationState.clearMultiselect) {
      setSelectedOptions([]);
      setFilters({});
      tableMutationDispatch({ type: TABLE_ACTIONS.UNSET_CLEAR_MULTISELECT });
    }
  }, [tableMutationState.clearMultiselect]);

  return (
    <ReactSelect
      isMulti={true}
      value={selectedOptions}
      options={options}
      placeholder={placeholder}
      onChange={values =>
        onChangeHandler(
          setSelectedOptions,
          filters,
          setFilters,
          filterType,
          values
        )
      }
      styles={ReactSelectStyles}
      {...props}
    />
  );
}

export default FilterMultiSelect;
