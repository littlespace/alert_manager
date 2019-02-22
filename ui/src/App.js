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
import { MuiThemeProvider, createMuiTheme } from '@material-ui/core/styles';
import theme from './theme';

/// ------------------------------------------------------
/// Views
/// ------------------------------------------------------
import AlertsListView from './views/AlertsListView';
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

  constructor(props, context) {
    super(props, context);
  }

  render() {
    return (
          <MuiThemeProvider theme={theme}>
            <div>
                <Header />
                <Switch>
                  <Route exact path="/login" component={SignIn} />
                  <PrivateRoute exact authed={Auth.loggedIn()} path="/" component={AlertsListView} />
                  <PrivateRoute exact authed={Auth.loggedIn()} path="/alerts" component={AlertsListView} />
                  <PrivateRoute exact authed={Auth.loggedIn()} path="/alert/:id/" component={AlertView} />
                  <PrivateRoute exact authed={Auth.loggedIn()} path="/suppression-rules" component={SuppressionRulesListView} />
                  <Route render={this._redirectToHome} />
                </Switch>
            </div>
          </MuiThemeProvider>
    );
  }
}
