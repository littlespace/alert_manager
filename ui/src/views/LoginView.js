import React, { useState, useRef } from "react";
import { Redirect } from "react-router-dom";
import styled from "styled-components";

import { AlertManagerApi } from "../library/AlertManagerApi";
import { SECONDARY } from "../styles/styles";

import Form from "../components/Forms/Form";

const api = new AlertManagerApi();

const FormContainer = styled.div`
  position: fixed;
  top: 20%;
  left: 50%;
  width: 600px;
  z-index: 10;
  transform: translate(-50%, 0);
`;

export default function LoginView() {
  const user = useRef();
  const password = useRef();
  const [authenticated, setAuthenticated] = useState(api.loggedIn());
  const [failedAuth, setFailedAuth] = useState(false);

  const authSuccess = () => {
    console.log("Successfully authenticate");
    setAuthenticated(true);
  };

  const authFailure = () => {
    console.log("Failed to authenticate");
    setAuthenticated(false);
    setFailedAuth(true);
  };

  const handleOnSubmit = event => {
    event.preventDefault();
    api.login(
      user.current.value,
      password.current.value,
      authSuccess,
      authFailure
    );
  };

  return (
    <>
      {authenticated ? (
        <Redirect to="/" />
      ) : (
        <FormContainer>
          <Form
            onSubmit={event => handleOnSubmit(event)}
            alert={failedAuth}
            backgroundColor={SECONDARY}
            canClose={false}
            title={"Alert Manager"}
          >
            <label>
              <span>Username</span>
              <input ref={user} required></input>
            </label>
            <label>
              <span>Password</span>
              <input type="password" ref={password} required></input>
            </label>
            <button type="submit">Sign In</button>
          </Form>
        </FormContainer>
      )}
      ;
    </>
  );
}
