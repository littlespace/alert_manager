

import React from 'react';
import PropTypes from 'prop-types';
import Select from '@material-ui/core/Select';
import Input from '@material-ui/core/Input';
import MenuItem from '@material-ui/core/MenuItem';
import FormControl from '@material-ui/core/FormControl';

let alertStatuses = [
  { label: "ACTIVE", id: 1 },
  { label: "SUPPRESSED", id: 2 },
  { label: "EXPIRED", id: 3 },
  { label: "CLEARED", id: 4 },
]

const SelectAlertStatusList = ({
  value,
  onChange,
  classe
}) => (
  <FormControl variant="outlined">
    <Select
            multiple
            value={value}
            onChange={onChange}
            input={<Input id="select-multiple" />}
            // className={classe}
          >
            {/* <MenuItem
                key="all"
                value="all">
                All
            </MenuItem> */}
            {alertStatuses.map(status => (
              <MenuItem
                key={status.id}
                value={status.id}
              >
                {status.label}
              </MenuItem>
            ))}
      </Select>
      </FormControl>

);

SelectAlertStatusList.propTypes = {
  onChange: PropTypes.func.isRequired,
  value: PropTypes.object.isRequired,
  classe: PropTypes.object.isRequired,
};

export default SelectAlertStatusList;
