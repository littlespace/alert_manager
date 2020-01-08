import React, { useState, useEffect, useContext } from "react";
import styled, { css } from "styled-components";
import { withRouter } from "react-router-dom";

import { NotificationContext } from "../components/contexts/NotificationContext";

import { AlertManagerApi } from "../library/AlertManagerApi";
import { HIGHLIGHT, SECONDARY } from "../styles/styles";
import { SEVERITY_COLORS } from "../library/utils";

import ActionsBar from "../components/Alerts/ActionsBar";
import AlertDetails from "../components/Alerts/AlertDetails";
import AlertHistory from "../components/Alerts/AlertHistory";
import AlertSpinner from "../components/Spinners/AlertSpinner";
import ContributingAlerts from "../components/Alerts/ContributingAlerts";
import NotificationBar from "../components/NotificationBar";
import SuppressForm from "../components/Forms/SuppressForm";

const api = new AlertManagerApi();

const Wrapper = styled.div`
  display: grid;
  grid-row-gap: 2.5em;
  color: ${HIGHLIGHT};
  overflow: hidden;

  & > div {
    margin-right: 1.5em;
    margin-left: 1.5em;
  }

  & > div:first-of-type {
    margin-top: 2em;
  }

  & > div:last-of-type {
    margin-bottom: 3em;
  }
`;

const ActionsBarContainer = styled.div`
  background-color: ${({ severity }) =>
    SEVERITY_COLORS[severity.toLowerCase()]["background-color"]};
  border-radius: 20px;
`;

const ContainerGlobalCss = css`
  background-color: ${SECONDARY};
  border-radius: 20px;
  padding: 0 1em 1em;
`;

const AlertDetailsContainer = styled.div`
  ${ContainerGlobalCss}
  max-width: 100%;
`;

const ContributingAlertsContainer = styled.div`
  ${ContainerGlobalCss}
  overflow: auto;
`;

const AlertHistoryContainer = styled.div`
  ${ContainerGlobalCss}
  height: 45vh;
  overflow: auto;
`;

const Title = styled.h2`
  background-color: ${SECONDARY};
  padding: 0.5em;
  position: sticky;
  margin: 0 0 1em;
  top: 0;
`;
const sortHistory = history => {
  history.sort((f, s) =>
    f.timestamp < s.timestamp ? 1 : f.timestamp > s.timestamp ? -1 : 0
  );
};

function AlertView(props) {
  const [alert, setAlert] = useState();
  const [loading, setLoading] = useState(true);
  const [showSuppressForm, setShowSuppressForm] = useState(false);
  const [updated, setUpdated] = useState(false);

  const {
    notificationBar,
    setNotificationBar,
    notificationColor,
    notificationMsg
  } = useContext(NotificationContext);

  useEffect(() => {
    const fetchAlert = async () => {
      const result = await api.getAlertWithHistory(props.match.params.id);

      result.relatedAlerts = await api.getContributingAlerts(
        props.match.params.id
      );

      sortHistory(result.history);
      setAlert(result);
      setLoading(false);
      setUpdated(false);
    };
    console.log("Fetching Alert Details");
    fetchAlert();
  }, [updated]);

  useEffect(() => {
    if (notificationBar === true) {
      setTimeout(() => setNotificationBar(false), 10000);
    }
  }, [notificationBar]);

  return (
    <>
      {notificationBar ? (
        <NotificationBar color={notificationColor} msg={notificationMsg} />
      ) : null}
      {showSuppressForm ? (
        <SuppressForm
          setShowSuppressForm={setShowSuppressForm}
          alert={alert}
          api={api}
          setUpdated={setUpdated}
        />
      ) : null}
      {loading ? (
        <AlertSpinner />
      ) : (
        <Wrapper>
          <ActionsBarContainer severity={alert.severity}>
            <ActionsBar
              alert={alert}
              api={api}
              setUpdated={setUpdated}
              setShowSuppressForm={setShowSuppressForm}
            />
          </ActionsBarContainer>
          <AlertDetailsContainer>
            <Title>{alert.name}</Title>
            <AlertDetails alert={alert} />
          </AlertDetailsContainer>
          {alert.relatedAlerts.length > 0 ? (
            <ContributingAlertsContainer>
              <Title>Contributing Alerts</Title>
              <ContributingAlerts alerts={alert.relatedAlerts} />
            </ContributingAlertsContainer>
          ) : null}
          <AlertHistoryContainer>
            <Title>Alert History</Title>
            <AlertHistory history={alert.history} />
          </AlertHistoryContainer>
        </Wrapper>
      )}
    </>
  );
}

export default withRouter(AlertView);
