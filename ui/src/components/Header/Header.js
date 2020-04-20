import React from "react";
import styled from "styled-components";

import AccountCircleIcon from "@material-ui/icons/AccountCircle";

import { AlertManagerApi } from "../../library/AlertManagerApi";
import { PagesDoc } from "../../static";
import { HIGHLIGHT, INFO, WARN, CRITICAL, ROBLOX } from "../../styles/styles";

const Auth = new AlertManagerApi();

function logout() {
  Auth.logout();
  window.location = "/login";
}

const Heading = styled.div`
  display: grid;
  grid-template-columns: repeat(3, auto);
  align-items: center;
  padding: 0.5em 1em;
  justify-content: space-between;
  background-color: ${ROBLOX};

  @media screen and (max-width: 850px) {
    justify-items: center;
    grid-template-columns: 1fr;
    grid-template-rows: repeat(3, auto);
  }
`;

const Title = styled.div`
  grid-column: 1;
  font-weight: 100;
  font-size: x-large;
  color: ${HIGHLIGHT};

  span {
    font-weight: 600;
  }
  @media screen and (max-width: 850px) {
    grid-column: 1;
    grid-row: 1;
  }
`;

const Menu = styled.div`
  display: flex;
  flex-wrap: wrap;
  white-space: nowrap;
  font-size: medium;
  font-weight: 400;
  grid-column: 2;

  @media screen and (max-width: 850px) {
    flex-direction: column;
    grid-column: 1;
    grid-row: 2;
  }
`;

const Link = styled.a`
  color: ${HIGHLIGHT};
  padding: 1em;
  text-decoration: none;

  :hover {
    animation: color-change 3s infinite;
  }

  @keyframes color-change {
    0% {
      color: ${INFO};
    }
    50% {
      color: ${WARN};
    }
    100% {
      color: ${CRITICAL};
    }
  }
`;

const LoginIcon = styled.div`
  grid-column: 3;

  @media screen and (max-width: 850px) {
    grid-column: 1;
    grid-row: 3;
  }
`;

const Logout = styled.button`
  background-color: ${ROBLOX};
  padding-left: 5px;
  position: relative;
  bottom: 11px;
  cursor: pointer;
  border: none;
  color: ${HIGHLIGHT};
  font-size: medium;
  font-weight: 400;
`;

function Header() {
  return (
    <Heading>
      <Title>
        Roblox<span>AlertManager</span>
      </Title>
      {Auth.loggedIn() === true ? (
        <>
          <Menu>
            <Link href={PagesDoc.alerts.url}>Alerts</Link>
            <Link href={PagesDoc.suppressionRules.url}>Suppressions</Link>
            <Link href={PagesDoc.users.url}>Users</Link>
          </Menu>
          <LoginIcon>
            <AccountCircleIcon fontSize="large" style={{ color: HIGHLIGHT }} />
            <Logout onClick={() => logout()}>Logout</Logout>
          </LoginIcon>
        </>
      ) : null}
    </Heading>
  );
}

export default Header;
