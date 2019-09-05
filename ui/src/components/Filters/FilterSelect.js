import React from 'react';
import styled from 'styled-components';

import { PRIMARY, HIGHLIGHT } from '../../styles/styles';

const TIME_RANGES_H = [
  { label: '1H', value: '1h' },
  { label: '24H', value: '24h' },
  { label: '7 Days', value: '168h' },
  { label: '14 Days', value: '336h' },
  { label: '30 Days', value: '720h' },
];

const Select = styled.select`
  position: absolute;
  top: 17px;
  right: 50px;
  background-color: ${PRIMARY};
  color: ${HIGHLIGHT};
  border: 0px solid ${PRIMARY};
  border-bottom: 1px solid ${HIGHLIGHT};
  font-size: medium;
  ::placeholder {
    color: ${HIGHLIGHT};
  }
`;

function FilterSelect({ ...props }) {
  return (
    <Select onChange={e => props.setTimeRange(e.target.value)}>
      <option value="" selected disabled hidden>
        Select Time Range
      </option>
      {TIME_RANGES_H.map(time => (
        <option value={time['value']}>{time['label']}</option>
      ))}
    </Select>
  );
}

export default FilterSelect;
