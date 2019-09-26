import React from "react";
import { CRITICAL, WARN } from "../styles/styles";
import styled from "styled-components";

const WarningBanner = styled.div`
  position: sticky;
  top: 0;
  background-color: ${props => (props.bgColor ? props.bgColor : WARN)};
  color: null;
  padding: 1em;
  text-align: center;
  z-index: 10;

  p {
    padding: 0;
    margin: 0;
  }
`;

const ALERT_MANAGER_SERVER = {
  prod: "https://alert-manager-api.simulprod.com/",
  integration: "https://alert-manager-integration-api.simulprod.com/"
};

export default function Warning({ NODE_ENV, REACT_APP_ALERT_MANAGER_SERVER }) {
  // Integration and Production deployments have "production" versions of node
  const deployedNodeEnv = NODE_ENV === "production" ? true : false;
  const prodBackend =
    REACT_APP_ALERT_MANAGER_SERVER === ALERT_MANAGER_SERVER.prod ? true : false;

  const warnMsg = (
    <p>
      <strong>!!WARNING!! Using Production Alert Database.</strong> Proceed with
      caution. Changes will affect our production alerts.
    </p>
  );
  const infoMsg = (
    <p>
      <strong>Using Test Alert Database.</strong> These alerts may not match{" "}
      <strong>production</strong> alerts and may include test data.
    </p>
  );

  /* Different Scenarios We Warn For
    1. Local Development: 
      - INFORM: When using the test (integration) backend
      - WARN: When using PROD alert backend
    2. Production (Integration and Production Deployments)
      - INFORM: WHen using the test (integration) backend */
  let warning = null;
  if (!deployedNodeEnv && prodBackend) {
    warning = <WarningBanner bgColor={CRITICAL}>{warnMsg}</WarningBanner>;
  } else if (!deployedNodeEnv && !prodBackend) {
    warning = <WarningBanner>{infoMsg}</WarningBanner>;
  } else if (deployedNodeEnv && !prodBackend) {
    warning = <WarningBanner>{infoMsg}</WarningBanner>;
  }

  return warning;
}
