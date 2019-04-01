

import React from 'react';
import PropTypes from 'prop-types';

import { withStyles } from '@material-ui/core/styles';
import MenuList from '@material-ui/core/MenuList';
import MenuItem from '@material-ui/core/MenuItem';
import ListItemText from '@material-ui/core/ListItemText'

import { Link } from 'react-router-dom'
import MenuItemLink from './MenuItemLink';

// -------------------------------------------------------
// Icons
// -------------------------------------------------------
import ListItemIcon from '@material-ui/core/ListItemIcon';
import ViewListIcon from '@material-ui/icons/ViewList';
import CodeIcon from '@material-ui/icons/Code';

// const styles = {
//     root: {
//         width: '200px',
//         overflowX: 'auto',
//     },
//     };

const drawerWidth = 200;

const styles = theme => ({
  root: {
    flexGrow: 1,
    height: 440,
    zIndex: 1,
    overflow: 'hidden',
    position: 'relative',
    display: 'flex',
  },
  appBar: {
    zIndex: theme.zIndex.drawer + 1,
  },
  drawerPaper: {
    position: 'relative',
    width: drawerWidth,
  },
  content: {
    flexGrow: 1,
    backgroundColor: theme.palette.background.default,
    padding: theme.spacing.unit * 3,
    minWidth: 0, // So the Typography noWrap works
  },
  toolbar: theme.mixins.toolbar,
  menuItem: {
    '&:focus': {
      backgroundColor: theme.palette.primary.main,
      '& $primary, & $icon': {
        color: theme.palette.common.white,
      },
    },
  },
  primary: {},
  icon: {
    marginRight: "0px"
  },
});

class Menu extends React.Component {

    static contextTypes = {
        router: PropTypes.object
    }

    constructor(props, context) {
        super(props, context);
        this.classes = this.props.classes;
    }

    render() {
        return (
        <div>
          <MenuList>
              <MenuItemLink className={this.classes.menuItem} to="/alerts">
                  <ListItemIcon className={this.classes.icon}>
                      <ViewListIcon />
                  </ListItemIcon>
                  <ListItemText classes={{ primary:this.classes.primary }} inset primary="Alerts" />
              </MenuItemLink>
              <MenuItemLink className={this.classes.menuItem} to="/suppression-rules">
                  <ListItemIcon className={this.classes.icon}>
                      <CodeIcon />
                  </ListItemIcon>
                  <ListItemText classes={{ primary:this.classes.primary }} inset primary="Suppression Rules" />
              </MenuItemLink>
              {/* <MenuItemLink className={this.classes.menuItem} to="/alertgrid">
                  <ListItemIcon className={this.classes.icon}>
                      <ViewListIcon />
                  </ListItemIcon>
                  <ListItemText classes={{ primary:this.classes.primary }} inset primary="Alerts Grid" />
              </MenuItemLink> */}
          </MenuList>
        </div>
        )};
}

export default withStyles(styles)(Menu);
