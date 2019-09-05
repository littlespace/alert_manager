import React from 'react';
import ReactSelect from 'react-select';

import {
  PRIMARY,
  SECONDARY,
  HIGHLIGHT,
  INFO,
  WARN,
  CRITICAL,
} from '../../styles/styles';

// Using ReactSelects custom way to style
// TODO: Any way we can figure out how to do styled components here?
const ReactSelectStyles = {
  control: styles => ({ ...styles, backgroundColor: SECONDARY }),
  option: (styles, state) => ({
    ...styles,
    color: state.isFocused ? HIGHLIGHT : PRIMARY,
    backgroundColor: state.isFocused ? PRIMARY : HIGHLIGHT,
  }),
  placeholder: styles => ({ ...styles, color: HIGHLIGHT, fontSize: '0.875em' }),
  input: styles => ({ ...styles, color: HIGHLIGHT }),
  dropdownIndictor: styles => ({ ...styles, color: HIGHLIGHT }),
  multiValue: styles => {
    return {
      ...styles,
      backgroundColor: PRIMARY,
      color: HIGHLIGHT,
      fontSize: '0.875em',
    };
  },
  multiValueLabel: (styles, { data }) => {
    let color = HIGHLIGHT;
    if (data.value.toLowerCase() === 'info') {
      color = INFO;
    } else if (data.value.toLowerCase() === 'warn') {
      color = WARN;
    } else if (data.value.toLowerCase() === 'critical') {
      color = CRITICAL;
    }
    return { ...styles, color: color };
  },
};

const onChangeHandler = (filters, setFilters, filterType, entries) => {
  // e.g if user deleted all the options, we will have no entries
  if (entries === null) {
    delete filters[filterType];
    setFilters(filters => ({ ...filters }));
  } else {
    let newEntries = [];
    entries.forEach(e => newEntries.push(e.value));
    setFilters(filters => ({ ...filters, [filterType]: newEntries }));
  }
};

function FilterMultiSelect({
  filters,
  setFilters,
  filterType,
  options,
  placeholder,
}) {
  return (
    <ReactSelect
      isMulti={true}
      options={options}
      placeholder={placeholder}
      onChange={entries =>
        onChangeHandler(filters, setFilters, filterType, entries)
      }
      styles={ReactSelectStyles}
    />
  );
}

export default FilterMultiSelect;
