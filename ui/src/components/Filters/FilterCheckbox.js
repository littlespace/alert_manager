import React from 'react';
import styled from 'styled-components';

const Checkbox = styled.input.attrs({ type: "checkbox" })`
  position: absolute;
  top: 17px;
  width: 17px;
  height: 17px;
`;

const Title = styled.span` 
  position: absolute;
  left: 53px;
  top: 19px;
`

function FilterCheckbox({ ...props }) {
  return (
    <>
    <Checkbox 
        checked={props.checked}
        onChange={props.onChangeHandler}
        />
    <Title>{props.title}</Title>
    </>
  )
}

export default FilterCheckbox;