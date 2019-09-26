import React, { useState } from "react";
import styled from "styled-components";

import { HIGHLIGHT, PRIMARY } from "../styles/styles";

const ToolTipBubbleTop = styled.div`
  position: absolute;
  z-index: 10;
  bottom: 100%;
  left: 50%;
  padding-bottom: 9px;
  transform: translateX(-50%);
  ::after {
    content: "";
    position: absolute;
    border-left: 9px solid transparent;
    border-right: 9px solid transparent;
    border-top: 9px solid
      ${props => (props.background ? props.background : HIGHLIGHT)};
    bottom: 0;
    left: 50%;
    transform: translateX(-50%);
  }
`;

const ToolTipBubbleLeft = styled.div`
  position: absolute;
  z-index: 10;
  top: 50%;
  right: 100%;
  padding-right: 9px;
  transform: translateY(-50%);
  ::after {
    content: "";
    position: absolute;
    border-left: 9px solid
      ${props => (props.background ? props.background : HIGHLIGHT)};
    border-top: 9px solid transparent;
    border-bottom: 9px solid transparent;
    top: 50%;
    right: 0;
    transform: translateY(-50%);
  }
`;

const Tooltip = styled.span`
  position: relative;
`;

const ToolTipTrigger = styled.span`
  display: inline-block;
  text-decoration: none;
  :hover {
    color: ${PRIMARY};
    font-weight: 500;
  }
`;

const ToolTipMessage = styled.div`
  background: ${props => (props.background ? props.background : HIGHLIGHT)};
  border-radius: 10px;
  color: ${props => (props.color ? props.color : PRIMARY)};
  font-size: 0.75rem;
  line-height: 1.4;
  padding: 1.5em 0.75em;
  min-width: 10vw;
  max-width: 40vw;
  text-align: center;
  box-shadow: 1px 2px 5px ${PRIMARY};
`;

function ToolTipBubble({ position, ...props }) {
  if (position === "top") {
    return <ToolTipBubbleTop {...props} />;
  } else if (position === "left") {
    return <ToolTipBubbleLeft {...props} />;
  }
}

export default function ToolTip({ msg, value, ...props }) {
  const [toolTip, setToolTip] = useState(false);

  return (
    <Tooltip onMouseLeave={() => setToolTip(false)}>
      {toolTip && (
        <ToolTipBubble {...props}>
          <ToolTipMessage {...props}>{msg}</ToolTipMessage>
        </ToolTipBubble>
      )}
      <ToolTipTrigger onMouseOver={() => setToolTip(true)}>
        {value}
      </ToolTipTrigger>
    </Tooltip>
  );
}
