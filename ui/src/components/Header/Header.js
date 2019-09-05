import React from 'react';
import styled from 'styled-components';

import AccountCircleIcon from '@material-ui/icons/AccountCircle';

import { AlertManagerApi } from '../../library/AlertManagerApi';
import { PagesDoc } from '../../static';
import { HIGHLIGHT, INFO, WARN, CRITICAL, ROBLOX } from '../../styles/styles';

const Auth = new AlertManagerApi();

function logout() {
  Auth.logout();
  window.location = '/login';
}

const Heading = styled.div`
  background-color: ${ROBLOX};
  border-bottom: 1px solid ${HIGHLIGHT};
  height: 50px;
`;

const Menu = styled.span`
  font-size: medium;
  font-weight: 400;
  position: absolute;
  left: 36%;
`;

const Link = styled.a`
  color: ${HIGHLIGHT};
  padding: 15px;
  float: left;
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

const Title = styled.text`
  font-weight: 100;
  font-size: x-large;
  color: ${HIGHLIGHT};
  position: absolute;
  top: 10px;
  left: 25px;

  span {
    font-weight: 600;
  }
`;

const LoginIcon = styled.span`
  float: right;
  margin-right: 20px;
  margin-top: 7px;
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
    <>
      <Heading>
        <Title>
          Roblox<span>AlertManager</span>
        </Title>
        <Menu>
          <Link href={PagesDoc.home.url}>Home</Link>
          <Link href={PagesDoc.alertsExplorer.url}>Alert Explorer</Link>
          <Link href={PagesDoc.ongoingAlerts.url}>Ongoing Alerts</Link>
          <Link href={PagesDoc.suppressionRules.url}>Suppressions</Link>
        </Menu>
        {Auth.loggedIn() === true ? (
          <LoginIcon>
            <AccountCircleIcon fontSize="large" style={{ color: HIGHLIGHT }} />
            <Logout onClick={() => logout()}>Logout</Logout>
          </LoginIcon>
        ) : null}
      </Heading>
    </>
  );
}

export default Header;
