

import React from 'react';
import PropTypes from 'prop-types';
import Select from '@material-ui/core/Select';
import Input from '@material-ui/core/Input';
import MenuItem from '@material-ui/core/MenuItem';

const devices = [
    'br1-ord1',
    'br2-ord1',
    'br1-sjc1',
    'br2-sjc1',
    'br1-iad1',
    'br2-iad1',
    'br1-lax1',
    'br2-lax1',
    'br1-lga1',
    'br2-lga1',
    'br1-lhr1',
    'br2-lhr1',
    'br1-fra1',
    'br2-fra1',
    'br1-hkg1',
    'br2-hkg1',
    'br1-nrt1',
    'br2-nrt1',
  ];


const SelectDevicesList = ({
  value,
  onChange,
  classe
}) => (
    <Select className={classe}
            multiple
            value={value}
            onChange={onChange}
            input={<Input id="select-multiple" />}
            // MenuProps={MenuProps}
          >
            {devices.map(name => (
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

SelectDevicesList.propTypes = {
  onChange: PropTypes.func.isRequired,
  value: PropTypes.object.isRequired,
  classe: PropTypes.object.isRequired,
};

export default SelectDevicesList;
