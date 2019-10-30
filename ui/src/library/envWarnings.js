import React from "react";
import { CRITICAL, PRIMARY, WARN } from "../styles/styles";
import styled from "styled-components";

const WarningBanner = styled.div`
  position: sticky;
  top: 0;
  background-color: ${props => (props.critical ? CRITICAL : WARN)};
  color: ${PRIMARY};
  padding: 0.6em;
  text-align: center;
  z-index: 10;

  p {
    padding: 0;
    margin: 0;
  }
`;

const ALERT_MANAGER_API = {
  prod: "https://alert-manager-api.simulprod.com/",
  integration: "https://alert-manager-integration-api.simulprod.com/"
};

const ALERT_MANAGER_INTEGRATION_SERVER =
  "https://alert-manager-integration.simulprod.com";

export default function Warning({ NODE_ENV, REACT_APP_ALERT_MANAGER_SERVER }) {
  // Integration and Production deployments have "production" versions of node
  const deployedNodeEnv = NODE_ENV === "production" ? true : false;
  const prodBackend =
    REACT_APP_ALERT_MANAGER_SERVER === ALERT_MANAGER_API.prod ? true : false;
  const integrationEnv =
    window.location.origin === ALERT_MANAGER_INTEGRATION_SERVER;

  const criticalMsg = (
    <>
      <p>
        <strong>!! WARNING !! Using Production Alert Database.</strong>
      </p>
      <p>Proceed with caution. Changes will affect our production alerts.</p>
    </>
  );
  const warnMsg = (
    <>
      <p>
        <strong>Using Test Alert Database.</strong>
      </p>
      <p>
        These alerts may not match <strong>production</strong> alerts and may
        include test data.
      </p>
    </>
  );

  /* Different Scenarios We Warn For
    1. Local Development: 
      - INFORM: When using the test (integration) backend
      - WARN: When using PROD alert backend
    2. Production (Integration and Production Deployments)
      - INFORM: WHen using the test (integration) backend 
      - WARN: When using Integration Server with PROD Backend (This should never happen) */
  let warning = null;
  if (!deployedNodeEnv && prodBackend) {
    warning = <WarningBanner critical>{criticalMsg}</WarningBanner>;
  } else if (!deployedNodeEnv && !prodBackend) {
    warning = <WarningBanner>{warnMsg}</WarningBanner>;
  } else if (deployedNodeEnv && !prodBackend) {
    warning = <WarningBanner>{warnMsg}</WarningBanner>;
  } else if (integrationEnv && prodBackend) {
    warning = <WarningBanner critical>{criticalMsg}</WarningBanner>;
  }

  return warning;
}
