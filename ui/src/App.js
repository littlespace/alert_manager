import React, { Component } from "react";
import { Router } from "react-router";
import { Redirect, Route, Switch } from "react-router-dom";
import PrivateRoute from "./components/PrivateRoute/PrivateRoute";
import createHistory from "history/createBrowserHistory";

/// ------------------------------------------------------
/// Theme
/// ------------------------------------------------------
import { MuiThemeProvider } from "@material-ui/core/styles";
import theme from "./theme";

/// ------------------------------------------------------
/// Views
/// ------------------------------------------------------
import HomeView from "./views/HomeView";
import OngoingAlertsView from "./views/OngoingAlertsView";
import AlertsExplorerView from "./views/AlertsExplorerView";
import AlertView from "./views/AlertView";
import AlertsView from "./views/AlertsView";
import SuppressionRulesListView from "./views/SuppRulesList";

/// ------------------------------------------------------
/// Menu & Login page
/// ------------------------------------------------------
import LoginView from "./views/LoginView";
import Header from "./components/Header/Header";
import { PagesDoc } from "./static";

import Warning from "./library/envWarnings";
import { AlertManagerApi } from "./library/AlertManagerApi";
import { NotificationProvider } from "./components/contexts/NotificationContext";

const history = createHistory();
const Auth = new AlertManagerApi();
const env = process.env;

export default class App extends Component {
  _redirectToHome() {
    return <Redirect to="/" />;
  }

  render() {
    return (
      <>
        <Warning {...env} />

        <Router history={history}>
          <MuiThemeProvider theme={theme}>
            <NotificationProvider>
              <div>
                <Header />
                <Switch>
                  <Route exact path="/login" component={LoginView} />
                  <PrivateRoute
                    exact
                    authed={Auth.checkToken()}
                    path="/"
                    component={AlertsView}
                  />
                  <PrivateRoute
                    exact
                    authed={Auth.checkToken()}
                    path={PagesDoc.ongoingAlerts.url}
                    component={OngoingAlertsView}
                  />
                  <PrivateRoute
                    exact
                    authed={Auth.checkToken()}
                    path={PagesDoc.alerts.url}
                    component={AlertsView}
                  />
                  <PrivateRoute
                    exact
                    authed={Auth.checkToken()}
                    path={PagesDoc.alertsExplorer.url}
                    component={AlertsExplorerView}
                  />
                  <PrivateRoute
                    exact
                    authed={Auth.checkToken()}
                    path="/alert/:id/"
                    component={AlertView}
                  />
                  <PrivateRoute
                    exact
                    authed={Auth.checkToken()}
                    path={PagesDoc.suppressionRules.url}
                    component={SuppressionRulesListView}
                  />
                  <Route render={this._redirectToHome} />
                </Switch>
              </div>
            </NotificationProvider>
          </MuiThemeProvider>
        </Router>
      </>
    );
  }
}
