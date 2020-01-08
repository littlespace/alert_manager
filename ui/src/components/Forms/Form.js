import React from "react";
import styled from "styled-components";

import { IoMdClose } from "react-icons/io";

import { SECONDARY, HIGHLIGHT, PRIMARY, ROBLOX } from "../../styles/styles";

const FormContainer = styled.div`
  width: 100%;
  padding: 1em 2em 2em;
  border-radius: 15px;
  color: ${HIGHLIGHT};
  background-color: ${({ backgroundColor }) =>
    backgroundColor ? backgroundColor : PRIMARY};
  box-shadow: 5px 5px 25px ${PRIMARY};
  border: ${({ alert }) => (alert ? "1px solid red" : null)};
`;

const FormStyled = styled.form`
  display: grid;
  grid-template-columns: max-content 1fr;
  grid-column-gap: 1em;
  grid-row-gap: 2em;
  align-items: center;
  margin: 2em 2em 1em;

  label {
    grid-column: span 2;
    display: grid;
    grid: inherit;
    grid-template-columns: 9em 1fr;
    grid-gap: inherit;
  }

  label > span {
    align-self: center;
    font-weight: 400;
  }

  * > input {
    padding: 0.5em;
    color: ${PRIMARY};
    font-size: medium;
    border-radius: 4px;
    border-style: none;
  }

  button {
    grid-column: 2;
    font-size: large;
    font-weight: 400;
    justify-self: end;
    padding: 0.5em;
    width: 110px;
    cursor: pointer;
    color: ${HIGHLIGHT};
    border-radius: 10px;
    background-color: transparent;
    border: 2px solid ${HIGHLIGHT};
    transition: 0.5s;
    :hover {
      border-color: ${ROBLOX};
      color: ${ROBLOX};
      border-radius: 15px;
    }
  }
`;

const Close = styled(IoMdClose)`
  position: absolute;
  top: 1em;
  right: 1em;
  font-size: 32px;
  border-radius: 100%;
  padding: 0.15em;
  cursor: pointer;
  transition: 0.5s;
  :hover {
    background-color: ${SECONDARY};
    color: ${ROBLOX};
  }
`;

export default function Form({
  children,
  showForm,
  title,
  backgroundColor,
  alert,
  canClose = true,
  ...props
}) {
  return (
    <FormContainer backgroundColor={backgroundColor} alert={alert}>
      <h3>{title}</h3>
      <hr />
      {canClose ? <Close onClick={() => showForm(false)} /> : null}
      <FormStyled {...props}>{children}</FormStyled>
    </FormContainer>
  );
}
