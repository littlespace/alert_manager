
import React from 'react';
import { withStyles } from '@material-ui/core/styles';
import AlertsTable from "../components/AlertsTable2";
import TopBar from "../components/TopBar"
import Menu from "../components/Menu"

const styles = theme => ({
    root: {
      flexGrow: 1,
    //   height: 440,
      zIndex: 1,
      overflow: 'hidden',
      position: 'relative',
      display: 'flex',
      height: '100%',
    },

    content: {
      flexGrow: 1,
      backgroundColor: theme.palette.background.default,
      padding: theme.spacing.unit * 3,
      minWidth: 0, // So the Typography noWrap works
    },
});
class AlertsListView extends React.Component {

    constructor(props){
        super(props);
        this.classes = this.props.classes;
    };

    render() {
        return (
            <div>
                <div>
                    <TopBar />
                </div>
                <div className={this.classes.root}>
                    <Menu />
                    <AlertsTable className={this.classes.content}/>
                </div>
            </div>
        );
    }
}


// function AlertsListView(props) {
//     const { classes } = props;

// }

export default withStyles(styles)(AlertsListView);
