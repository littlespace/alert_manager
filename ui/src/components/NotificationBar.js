import React from "react";
import styled from "styled-components";

import NotificationsNoneOutlinedIcon from "@material-ui/icons/NotificationsNoneOutlined";

import { PRIMARY } from "../styles/styles";

const Container = styled.div`
  display: flex;
  position: fixed;
  z-index: 100;
  top: 1em;
  right: 1em;
  padding: 0.5em;
  background-color: ${props => (props.color ? props.color : PRIMARY)};
  border-radius: 3px;
  box-shadow: 1px 1px 5px 1px ${PRIMARY};
  font-size: small;
  color: ${PRIMARY};
`;
const Icon = styled.span`
  margin: auto;
`;
const Notification = styled.div`
  margin: auto;
`;

export default function NotificationBar({ color, msg }) {
  return (
    <Container color={color}>
      <Icon>
        <NotificationsNoneOutlinedIcon />
      </Icon>
      <Notification>{msg}</Notification>
    </Container>
  );
}
