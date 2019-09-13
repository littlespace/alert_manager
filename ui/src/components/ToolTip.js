import React, { useState } from 'react';
import styled from 'styled-components';

import { SECONDARY, HIGHLIGHT, PRIMARY } from "../styles/styles";

const ToolTipBubble = styled.div`
  position: absolute;
  z-index: 10;
  bottom: 100%;
  left: 50%;
  padding-bottom: 9px;
  transform: translateX(-50%);
<<<<<<< HEAD
  ::after {
    content: "";
    position: absolute;
    border-left: 9px solid transparent;
    border-right: 9px solid transparent;
    border-top: 9px solid ${props => props.background ? props.background : HIGHLIGHT};
    bottom: 0;
    left: 50%;
    transform: translateX(-50%);
  }
=======
>>>>>>> 7cb9e2d704d0f04a1f4fe3f56afc87d9828336e0
`;

const Tooltip = styled.span`
  position: relative;
`

const ToolTipTrigger = styled.span`
  display: inline-block;
  text-decoration: none;
  :hover {
    color: ${PRIMARY};
    font-weight: 500;
  }
`

const ToolTipMessage = styled.div`
<<<<<<< HEAD
  background: ${props => (props.background ? props.background : HIGHLIGHT)};
  border-radius: 15px;
  color: ${props => (props.color ? props.color : PRIMARY)};
  font-size: 0.75rem;
  line-height: 1.4;
  padding: 0.75em;
=======
  background: ${props => props.background ? props.background: HIGHLIGHT};
  border-radius: 15px;
  color: ${props => props.color ? props.color: PRIMARY};
  font-size: .75rem;
  line-height: 1.4;
  padding: .75em;
>>>>>>> 7cb9e2d704d0f04a1f4fe3f56afc87d9828336e0
  min-width: 10vw;
  max-width: 40vw;
  text-align: center;
  box-shadow: 1px 1px 5px ${SECONDARY};
<<<<<<< HEAD
`;
=======
`
>>>>>>> 7cb9e2d704d0f04a1f4fe3f56afc87d9828336e0

export default function ToolTip({ msg, value, ...props }) {
  const [ toolTip, setToolTip ] = useState(false)

  return (
    <Tooltip onMouseLeave={() => setToolTip(false)}>
      {toolTip && (
<<<<<<< HEAD
        <ToolTipBubble {...props}>
=======
        <ToolTipBubble>
>>>>>>> 7cb9e2d704d0f04a1f4fe3f56afc87d9828336e0
          <ToolTipMessage {...props}>{msg}</ToolTipMessage>
        </ToolTipBubble>
      )}
      <ToolTipTrigger onMouseOver={() => setToolTip(true)}>
        {value}
      </ToolTipTrigger>
    </Tooltip>
  );
}