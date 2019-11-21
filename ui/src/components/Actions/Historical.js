import React from "react";
import styled from "styled-components";

const Checkbox = styled.input.attrs({ type: "checkbox" })`
  margin: auto 0.5em auto;
  align-self: flex-start;
`;

const Title = styled.span`
  margin: auto auto auto 0;
  align-self: flex-start;
`;

export default function Historical({ ...props }) {
  return (
    <>
      <Checkbox checked={props.checked} onChange={props.onChangeHandler} />
      <Title>{props.title}</Title>
    </>
  );
}
