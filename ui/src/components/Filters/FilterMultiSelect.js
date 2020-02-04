import React, { useState, useEffect } from "react";
import ReactSelect from "react-select";

import { AlertManagerApi } from "../../library/AlertManagerApi";
import { useHistory, useLocation } from "react-router-dom";

import {
  CRITICAL,
  HIGHLIGHT,
  INFO,
  PRIMARY,
  SECONDARY,
  WARN
} from "../../styles/styles";

import {
  getFilterValuesFromType,
  getSearchOptionsByKey,
  setSearchString
} from "../../library/utils";

const api = new AlertManagerApi();

const TRANSLATE_FILTERS = ["status", "severity"];

// TODO:  convert to styled components
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

function getSelectedOptions(filterType, location) {
  let options = getSearchOptionsByKey(filterType, location);
  // TODO: Remove when we move "status" and "severity" into a DB table
  if (TRANSLATE_FILTERS.includes(filterType.toLowerCase())) {
    return getFilterValues(filterType, options);
  } else {
    return options.map(option => ({ label: option, value: option }));
  }
}

function getFilterValues(filterType, options) {
  // This expects to return string values for the given type
  const values = getFilterValuesFromType(filterType);
  // This will return an arry of options for the select, e.b [{ label: "warn", value: 1}, {...}]
  return options.map(option => ({ label: values[option], value: option }));
}

function FilterMultiSelect({ filterType, placeholder }) {
  const [options, setOptions] = useState([]);
  const [loading, setLoading] = useState(false);

  let history = useHistory();
  let location = useLocation();

  useEffect(() => {
    // Get the options and set them for the multi-select based on the select type.
    const getOptions = async () => {
      let options = await api.getDistinctField(filterType);
      // TODO: Remove when we move "status" and "severity" into a DB table
      if (TRANSLATE_FILTERS.includes(filterType.toLowerCase())) {
        setOptions(getFilterValues(filterType, options));
      } else {
        setOptions(options.map(option => ({ label: option, value: option })));
      }
    };
    setLoading(true);
    getOptions();
    setLoading(false);
  }, []);

  return (
    <ReactSelect
      isMulti
      value={getSelectedOptions(filterType, location)}
      options={options}
      placeholder={placeholder}
      onChange={values =>
        setSearchString(filterType, values, location, history)
      }
      styles={ReactSelectStyles}
      isLoading={loading}
    />
  );
}

export default FilterMultiSelect;
