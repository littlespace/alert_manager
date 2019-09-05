import React from 'react';
import styled from 'styled-components';

import { PRIMARY, HIGHLIGHT } from '../../styles/styles';

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

const onChangeHandler = (filters, setFilters, value, type) => {
  setFilters(filters => ({ ...filters, [type]: value }));
};

function FilterInput({ filters, setFilters, filterType, placeholder }) {
  return (
    <Input
      value={filters['filterType'] || undefined}
      onChange={e =>
        onChangeHandler(filters, setFilters, e.target.value, filterType)
      }
      placeholder={placeholder}
    ></Input>
  );
}

export default FilterInput;
