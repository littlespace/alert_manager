import React from 'react';

const SessionContext = React.createContext();

class SessionProvider extends React.Component {
    
    state = { 
      isAuth: false,
      username: null,
    }

    constructor() {
        super()
        this.login = this.login.bind(this)
        this.logout = this.logout.bind(this)
    }

    login() {
        // setting timeout to mimic an async login
        setTimeout(() => this.setState({ isAuth: true }), 1000)
        
    }

    logout() {
        this.setState({ isAuth: false })
    }
        
    render() {
      return (
        <SessionContext.Provider
          value={{ 
            isAuth: this.state.isAuth,
            login: this.login,
            logout: this.logout,
            username: this.state.username
          }}
        >
          {this.props.children}
        </SessionContext.Provider>
      )
    }
  }
  const SessionConsumer = SessionContext.Consumer
  export { SessionProvider, SessionConsumer }
