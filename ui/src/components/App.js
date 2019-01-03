import React, { Component } from 'react';
import {
  Redirect,
  Route,
  Switch,
} from 'react-router-dom';

import AlertsListView from '../views/AlertsListView';
import AlertView from '../views/AlertView';

import red from '@material-ui/core/colors/red';
import { MuiThemeProvider, createMuiTheme } from '@material-ui/core/styles';
import theme from '../theme';

export default class App extends Component {
  _redirectToHome() {
    return <Redirect to="/" />;
  }

  render() {
    return (
          <MuiThemeProvider theme={theme}>
            <div>
              {/* <Header /> */}
              {/* content */}
              <Switch>
                <Route exact path="/" component={AlertsListView} />
                <Route exact path="/alerts" component={AlertsListView} />
                <Route path="/alert/:id/" component={AlertView} />
                {/* catch-all redirects to home */}
                <Route render={this._redirectToHome} />
              </Switch>
            </div>
          </MuiThemeProvider>
    );
  }
}
