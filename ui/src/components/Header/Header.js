

import {
    withStyles
  } from "@material-ui/core/styles";

import React from 'react';
import Typography from '@material-ui/core/Typography';
import classNames from 'classnames';
import { Link } from 'react-router-dom'

import Button from '@material-ui/core/Button';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import AccountCircle from '@material-ui/icons/AccountCircle';

import { AlertManagerApi } from '../../library/AlertManagerApi';

import Tooltip from '@material-ui/core/Tooltip';

import Menu from '../Menu';

// ------------------------------------------------------------
// Icons
// ------------------------------------------------------------
import IconButton from '@material-ui/core/IconButton';
import MenuIcon from '@material-ui/icons/Menu';
import ChevronLeftIcon from '@material-ui/icons/ChevronLeft';

import Divider from '@material-ui/core/Divider';

import { SwipeableDrawer } from "@material-ui/core";


const Auth = new AlertManagerApi()
const drawerWidth = 240

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
    menuButton: {
        marginLeft: 12,
        marginRight: 20,
    },
    hide: {
        display: 'none',
    },
    drawerHeader: {
        display: 'flex',
        alignItems: 'center',
        padding: '0 8px',
        ...theme.mixins.toolbar,
        justifyContent: 'flex-end',
    },
    drawer: {
        width: drawerWidth,
        flexShrink: 0,
    },
    drawerPaper: {
        width: drawerWidth,
    },
});

function logout(event) {
    Auth.logout()
};

class Header extends React.Component {
// function Header(props) {
    constructor(props){
        super(props);
        this.classes = this.props.classes;
        this.state = {
            drawer_open: false
        }
    }

    toggleDrawer = (open) => () => {
        this.setState({
            drawer_open: open,
        });
    };

    render() {
        const open = this.state.drawer_open;

        return (
            <div  className={this.classes.root}> 
                <AppBar position="static">
                    <Toolbar>
                    <IconButton
                        color="inherit"
                        aria-label="Open drawer"
                        onClick={this.toggleDrawer(true)}
                        className={classNames(this.classes.menuButton, open && this.classes.hide)}
                    >
                    <MenuIcon />
                    </IconButton>
                        <Typography variant="h6" color="inherit" className={this.classes.grow} >
                            Roblox | Alert Manager
                        </Typography>
                        
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
                                    className={this.classes.button} 
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
                <SwipeableDrawer
                    className={this.classes.drawer}
                    variant="persistent"
                    anchor="left"
                    open={open}
                    onClose={this.toggleDrawer(false)}
                    onOpen={this.toggleDrawer(true)}

                    classes={{
                        paper: this.classes.drawerPaper,
                    }}
                    >
                    
                    <div className={this.classes.drawerHeader}>
                        <IconButton onClick={this.toggleDrawer(false)}>
                        <ChevronLeftIcon />
                        </IconButton>
                    </div>
                    <Divider />
                    <div
                        tabIndex={0}
                        role="button"
                        onClick={this.toggleDrawer(false)}
                        onKeyDown={this.toggleDrawer(false)}
                    >
                    <Menu/>
                    </div>
                </SwipeableDrawer>
            </div>
        )
    }
}

export default withStyles(styles)(Header);
