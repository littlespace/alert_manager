import React, { Component } from 'react';
import {
  Redirect,
  Route,
  Switch,
} from 'react-router-dom';
import PrivateRoute from './components/PrivateRoute/PrivateRoute'

/// ------------------------------------------------------
/// Theme
/// ------------------------------------------------------
import { MuiThemeProvider } from '@material-ui/core/styles';
import theme from './theme';

/// ------------------------------------------------------
/// Views
/// ------------------------------------------------------
import OngoingAlertsView from './views/OngoingAlertsView';
import AlertsExplorerView from'./views/AlertsExplorerView';
import AlertView from './views/AlertView';
import SuppressionRulesListView from './views/SuppRulesList';

/// ------------------------------------------------------
/// Menu & Login page
/// ------------------------------------------------------
import SignIn from './components/SignIn/SignIn'
import Header from './components/Header/Header';

import { AlertManagerApi } from './library/AlertManagerApi';

const Auth = new AlertManagerApi()

export default class App extends Component {
  
  _redirectToHome() {
    return <Redirect to="/" />;
  }

  render() {
    return (
          <MuiThemeProvider theme={theme}>
            <div>
                <Header/>
                <Switch>
                  <Route exact path="/login" component={SignIn} />
                  <PrivateRoute exact authed={Auth.checkToken()} path="/" component={OngoingAlertsView} />
                  <PrivateRoute exact authed={Auth.checkToken()} path="/ongoing-alerts" component={OngoingAlertsView} />
                  <PrivateRoute exact authed={Auth.checkToken()} path="/alerts-explorer" component={AlertsExplorerView} />
                  <PrivateRoute exact authed={Auth.checkToken()} path="/alert/:id/" component={AlertView} />
                  <PrivateRoute exact authed={Auth.checkToken()} path="/suppression-rules" component={SuppressionRulesListView} />
                  <Route render={this._redirectToHome} />
                </Switch>
            </div>
          </MuiThemeProvider>
    );
  }
}
