import React from "react";
import styled from "styled-components";

import { HIGHLIGHT, INFO, WARN, CRITICAL } from "../../styles/styles";

const CircleSpinner = styled.div`
  margin: auto auto 20px;
  border: 5px solid ${HIGHLIGHT};
  border-top: 5px solid ${INFO};
  border-right: 5px solid ${WARN};
  border-bottom: 5px solid ${CRITICAL};
  border-radius: 50%;
  width: ${p => `${p.size}${p.sizeUnit}`};
  height: ${p => `${p.size}${p.sizeUnit}`};
  -webkit-animation: spin 1s linear infinite;
  animation: spin 1s linear infinite;
  @-webkit-keyframes spin {
    0% {
      -webkit-transform: rotate(0deg);
    }
    100% {
      -webkit-transform: rotate(360deg);
    }
  }
  @keyframes spin {
    0% {
      transform: rotate(0deg);
    }
    100% {
      transform: rotate(360deg);
    }
  }
`;

const Spinner = ({ size, sizeUnit }) => (
  <CircleSpinner size={size} sizeUnit={sizeUnit} />
);

Spinner.defaultProps = {
  size: 40,
  sizeUnit: "px"
};

export default Spinner;
