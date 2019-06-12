import React from 'react';

import Avatar from '@material-ui/core/Avatar';
import Button from '@material-ui/core/Button';
import FormControl from '@material-ui/core/FormControl';
import Input from '@material-ui/core/Input';
import InputLabel from '@material-ui/core/InputLabel';

import Paper from '@material-ui/core/Paper';
import Typography from '@material-ui/core/Typography';
import withStyles from '@material-ui/core/styles/withStyles';
import Snackbar from '@material-ui/core/Snackbar';
import SnackbarContent from '@material-ui/core/SnackbarContent';
import { Redirect } from 'react-router-dom';

/// -------------------------------------
/// Icons 
/// -------------------------------------
import InfoIcon from '@material-ui/icons/Info';
import CloseIcon from '@material-ui/icons/Close';
import IconButton from '@material-ui/core/IconButton';
import LockOutlinedIcon from '@material-ui/icons/LockOutlined';

import { AlertManagerApi } from '../../library/AlertManagerApi';

const Auth = new AlertManagerApi()

const styles = theme => ({
  main: {
    width: 'auto',
    display: 'block', // Fix IE 11 issue.
    marginLeft: theme.spacing.unit * 3,
    marginRight: theme.spacing.unit * 3,
    [theme.breakpoints.up(400 + theme.spacing.unit * 3 * 2)]: {
      width: 400,
      marginLeft: 'auto',
      marginRight: 'auto',
    },
  },
  paper: {
    marginTop: theme.spacing.unit * 8,
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
    padding: `${theme.spacing.unit * 2}px ${theme.spacing.unit * 3}px ${theme.spacing.unit * 3}px`,
  },
  avatar: {
    margin: theme.spacing.unit,
    backgroundColor: theme.palette.secondary.main,
  },
  form: {
    width: '100%', // Fix IE 11 issue.
    marginTop: theme.spacing.unit,
  },
  submit: {
    marginTop: theme.spacing.unit * 3,
  },
});

class SignIn extends React.Component {

  constructor(props){
    super(props);
    this.classes = this.props.classes;
    this.authSuccessfull = this.authSuccessfull.bind(this);
    this.state = {
      username: null,
      password: null,
      redirect: false
    }
  }

  handleChange = name => event => {
    this.setState({ [name]: event.target.value });
  };

  login = () => {
    
    Auth.login(
        this.state.username, 
        this.state.password,
        this.authSuccessfull
    )
  }

  authSuccessfull() {
    this.setState({
      redirect: true
    })
  }

  handleMessageClose = (event, reason) => {
    if (reason === 'clickaway') {
      return;
    }
    this.setState({ snackbarUpdateMessage: false });
  };

  showSuccessMessage() {
      this.setState({ snackbarUpdateMessage: true });
  };

  renderRedirect = () => {
    if (this.state.redirect) {
      return <Redirect to='/' />
    } else {

    }
  }

  render() {
    return (
        <main className={this.classes.main}>
          <Paper className={this.classes.paper}>
            <Avatar className={this.classes.avatar}>
              <LockOutlinedIcon />
            </Avatar>
            <Typography component="h1" variant="h5">
              Sign in
            </Typography>
            <div className={this.classes.form}>
              {this.renderRedirect()}
              <FormControl margin="normal" required fullWidth>
                <InputLabel htmlFor="username">Username</InputLabel>
                <Input 
                  id="username" 
                  name="username" 
                  autoComplete="username" 
                  autoFocus
                  onChange={this.handleChange('username')} />
              </FormControl>
              <FormControl margin="normal" required fullWidth>
                <InputLabel htmlFor="password">Password</InputLabel>
                <Input 
                  name="password" 
                  type="password" 
                  id="password" 
                  autoComplete="current-password"
                  onChange={this.handleChange('password')} />
              </FormControl>
              <Button
                type="submit"
                fullWidth
                variant="contained"
                color="primary"
                className={this.classes.submit}
                onClick={this.login}
              >
                Sign in
              </Button>
            </div>
          </Paper>
          <Snackbar
                anchorOrigin={{
                    vertical: 'bottom',
                    horizontal: 'left',
                }}
                open={this.state.snackbarUpdateMessage}
                autoHideDuration={6000}
                onClose={this.handleMessageClose}
                >
                <SnackbarContent
                    className={this.classes.info}
                    aria-describedby="client-snackbar"
                    message={
                        <span id="client-snackbar" className={this.classes.message}>
                        <InfoIcon />
                        {'Alert Succefully updated'}
                        </span>
                    }
                    action={[
                        <IconButton
                            key="close"
                            aria-label="Close"
                            color="inherit"
                            className={this.classes.close}
                            onClick={this.handleMessageClose}
                        >
                        <CloseIcon className={this.classes.icon} />
                        </IconButton>,
                    ]}
                    />
            </Snackbar>
      </main>
    );
  }
}


export default withStyles(styles)(SignIn);