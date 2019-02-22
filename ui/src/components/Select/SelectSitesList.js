

import React from 'react';
import PropTypes from 'prop-types';
import Select from '@material-ui/core/Select';
import Input from '@material-ui/core/Input';
import MenuItem from '@material-ui/core/MenuItem';

const sites = [
    'ord1',
    'iad1',
    'sjc1',
    'lax1',
    'atl1',
    'lhr1',
    'fra1',
    'hkg1',
    'nrt1',
    'rp',
    'ash1',
    'chi1',
  ];

const SelectSitesList = ({
  value,
  onChange,
  classe
}) => (
    <Select
            multiple
            value={value}
            onChange={onChange}
            input={<Input id="select-multiple" />}
            className={classe}
            // MenuProps={MenuProps}
          >
            {/* <MenuItem
                key="all"
                value="all">
                All
            </MenuItem> */}
            {sites.map(name => (
              <MenuItem
                key={name}
                value={name}
                // style={{
                //   fontWeight:
                //     this.state.name.indexOf(name) === -1
                //       ? theme.typography.fontWeightRegular
                //       : theme.typography.fontWeightMedium,
                // }}
              >
                {name}
              </MenuItem>
            ))}
      </Select>
);

SelectSitesList.propTypes = {
//   onSubmit: PropTypes.func.isRequired,
  onChange: PropTypes.func.isRequired,
  value: PropTypes.object.isRequired,
  classe: PropTypes.object.isRequired,
//   onChange: PropTypes.func.isRequired,
//   errors: PropTypes.object.isRequired,
//   successMessage: PropTypes.string.isRequired,
//   user: PropTypes.object.isRequired
};

export default SelectSitesList;
