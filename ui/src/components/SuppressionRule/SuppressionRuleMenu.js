import React from 'react';

import Menu from '@material-ui/core/Menu';
import MenuItem from '@material-ui/core/MenuItem';
import MoreVertIcon from '@material-ui/icons/MoreVert';
import IconButton from '@material-ui/core/IconButton';


class SuppressionRuleMenu extends React.Component {

    constructor(props) {
        super(props);
        this.state = {
            anchorEl: null,
        };
    }

    handleClick = event => {
        event.preventDefault()
        this.setState({ anchorEl: event.currentTarget });
    };

    handleClose = () => {
        this.setState({ anchorEl: null });
    };

    handleDelete = () => {
        this.handleClose()
        this.props.onDelete()
    }

    render() {
        const { anchorEl } = this.state;
        return (
            <div>
                <IconButton
                    onClick={this.handleClick}
                >
                    <MoreVertIcon />
                </IconButton>
                <Menu
                    id="simple-menu"
                    anchorEl={anchorEl}
                    open={Boolean(anchorEl)}
                    onClose={this.handleClose}
                >
                    <MenuItem disabled={this.props.disabled} onClick={this.handleDelete}>Clear SuppRule</MenuItem>
                </Menu>
            </div>
        );
    }
}

export default SuppressionRuleMenu;