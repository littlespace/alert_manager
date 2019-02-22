
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import {
    createMuiTheme,
    MuiThemeProvider,
    withStyles
  } from "@material-ui/core/styles";

import React from 'react';
import Typography from '@material-ui/core/Typography';

import { Link } from 'react-router-dom'

import Button from '@material-ui/core/Button';
import ViewListIcon from '@material-ui/icons/ViewList';

import IconButton from '@material-ui/core/IconButton';
import AccountCircle from '@material-ui/icons/AccountCircle';
import MenuItem from '@material-ui/core/MenuItem';
import Menu from '@material-ui/core/Menu';

import { AlertManagerApi } from '../../library/AlertManagerApi';

import Tooltip from '@material-ui/core/Tooltip';


const Auth = new AlertManagerApi()

const styles = theme => ({
    leftIcon: {
        marginRight: theme.spacing.unit,
    },
    button: {
        margin: theme.spacing.unit,
    },
    root: {
        flexGrow: 1,
    },
    grow: {
        flexGrow: 1,
    },
});

function logout(event) {
    Auth.logout()
};

function Header(props) {
    const { classes } = props;
    return (
        <div  className={classes.root}> 
            <AppBar position="static">
                <Toolbar>
                    <Typography variant="title" color="inherit" className={classes.grow} >
                        Roblox | Alert Manager
                    </Typography>
                    {/* <Button 
                        color="inherit" 
                        className={classes.button } 
                        component={Link}
                        to="/alerts">
                        Alerts
                    </Button> */}

                    
                    {Auth.loggedIn() === true ?
                        <div>
                             <Tooltip title={Auth.getUsername()} placement="bottom">
                                <IconButton
                                    // aria-owns={open ? 'menu-appbar' : undefined}
                                    // aria-haspopup="true"
                                    // onClick={this.handleMenu}
                                    color="inherit"
                                >
                                <AccountCircle />
                                </IconButton>
                            </Tooltip>
                            <Button 
                                color="inherit" 
                                className={classes.button} 
                                component={Link}
                                to="/login"
                                onClick={logout}
                                >
                                Logout
                            </Button>

                            {/* <Menu
                            id="menu-appbar"
                            anchorEl={anchorEl}
                            anchorOrigin={{
                                vertical: 'top',
                                horizontal: 'right',
                            }}
                            transformOrigin={{
                                vertical: 'top',
                                horizontal: 'right',
                            }}
                            open={open}
                            onClose={this.handleClose}
                            >
                            {/* <MenuItem onClick={this.handleClose}>Profile</MenuItem>
                            <MenuItem onClick={this.handleClose}>My account</MenuItem> */}
                            {/* </Menu> */}
                        </div>
                         : 
                            <div>
                                {/* <Button color="inherit">Login</Button> */}
                            </div>
                        }

                    

                </Toolbar>
            </AppBar>
            
        </div>
    );
}

export default withStyles(styles)(Header);
