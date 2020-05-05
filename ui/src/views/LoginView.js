import React, { useContext, useEffect, useState, useRef } from "react";
import { Redirect } from "react-router-dom";
import styled from "styled-components";

import { NotificationContext } from "../components/contexts/NotificationContext";

import { AlertManagerApi } from "../library/AlertManagerApi";
import NotificationBar from "../components/Notifications/NotificationBar";
import { SECONDARY, CRITICAL } from "../styles/styles";

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

  const {
    notificationBar,
    notificationMsg,
    setNotificationBar,
    setNotificationMsg
  } = useContext(NotificationContext);

  const authFailure = response => {
    console.error(
      `HTTP ${response.status}::${response.statusText}::${response.statusMsg}`
    );
    let msg;
    if (response.status === 401) {
      // If string not found it will return -1. We want to check the difference if it's an AUTH failure
      // or just a network connection problem, which most likely is VPN connection.
      if (response.statusMsg.search("failed to connect to ldap") === -1) {
        msg = "Failed to authenticate to LDAP. Check username/password.";
      } else {
        msg = "LDAP server not found. Check VPN connection.";
      }
    } else if (response.status === 400) {
      msg = "User needs to be added to Alert Manager";
    } else {
      msg = "Unknown authentication error";
    }

    setAuthenticated(false);
    setFailedAuth(true);
    setNotificationMsg(`ERROR: ${msg}`);
    setNotificationBar(true);
  };

  const handleOnSubmit = async event => {
    event.preventDefault();
    let response = await api.login(user.current.value, password.current.value);
    if (response.ok) {
      console.log("Successfully authenticate");
      setAuthenticated(true);
    } else {
      authFailure(response);
    }
  };

  // Used for our notifications bar
  useEffect(() => {
    if (notificationBar === true) {
      setTimeout(() => setNotificationBar(false), 10000);
    }
  }, [notificationBar]);

  return (
    <>
      {authenticated ? (
        <Redirect to="/" />
      ) : (
        <>
          {notificationBar ? (
            <NotificationBar color={CRITICAL} msg={notificationMsg} />
          ) : null}
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
        </>
      )}
      ;
    </>
  );
}
