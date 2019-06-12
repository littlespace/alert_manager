

import React from 'react';
import {
    Redirect,
    Route,
  } from 'react-router-dom';

import { AlertManagerApi } from '../../library/AlertManagerApi';

const Auth = new AlertManagerApi()

function PrivateRoute ({component: Component, ...rest}) {
    
    return (
        <Route
          {...rest}
          render={(props) => Auth.loggedIn() === true
            ? <Component {...props} />
            : <Redirect to={{pathname: '/login', state: {from: props.location}}} />}
        />
    )
  }

export default PrivateRoute;

  