
import React from 'react';
import { withStyles } from '@material-ui/core/styles';

import Typography from '@material-ui/core/Typography';

import { AlertManagerApi } from '../library/AlertManagerApi';

import Card from '@material-ui/core/Card';
import CardActions from '@material-ui/core/CardActions';
import CardContent from '@material-ui/core/CardContent';
import Button from '@material-ui/core/Button';

import { Link } from 'react-router-dom'

import { PagesDoc } from './../static'

const styles = theme => ({
    root: {
      flexGrow: 1,
      zIndex: 1,
      overflow: 'hidden',
      position: 'relative',
      display: 'flex',
      height: '100%',
    },
    gridroot: {
        flexWrap: "nowrap",
        margin: "10px"
    },
    content: {
      flexGrow: 1,
      backgroundColor: theme.palette.background.default,
      padding: theme.spacing.unit * 3,
      minWidth: 0, // So the Typography noWrap works
    },
    pageTitle:{
        height: "30px",
        lineHeight: "30px",
        paddingLeft: "15px",
        paddingTop: "10px"
      },
    // content: {
    //     flexGrow: 1,
    //     padding: theme.spacing.unit * 3,
    //     height: '100vh',
    //     overflow: 'auto',
    //   },
      tableContainer: {
        height: "90%",
        display: 'flex',
      },
      card: {
        minWidth: 275,
        width: 400,
        height: 250,
        margin: 15
      },
      bullet: {
        display: 'inline-block',
        margin: '0 2px',
        transform: 'scale(0.8)',
      },
      title: {
        fontSize: 14,
      },
      pos: {
        marginBottom: 12,
      }
});

class HomeView extends React.Component {

    constructor(props){
        super(props);
        this.classes = this.props.classes;
        this.api = new AlertManagerApi();
        this.state = {
            NbrActive: 0,
            NbrSuppressed: 0, 
        }
    };

    componentDidMount(){
        this.updateAlertsList()
    }


    updateAlertsList = () => {
        this.api.getAlertsList({status: [1,2]})
            .then(data => this.processAlertsList(data));
    }

    processAlertsList(data) {

        let NbrActive = 0
        let NbrSuppressed = 0

        for(var i in data) {

            // Ignore all sites that are not listed in sites_location
            if (data[i].Status === "ACTIVE") {
                NbrActive++;
            } else if (data[i].Status === "SUPPRESSED") {
                NbrSuppressed++;
            } 
        }

        this.setState({ 
            NbrActive: NbrActive,
            NbrSuppressed: NbrSuppressed
         })

    }

    render() {
        return (
        <div>
            <div className={this.classes.tableContainer}>
            <Card className={this.classes.card}>
              <CardContent>
                <Typography variant="h5" component="h2">
                    {PagesDoc.ongoingAlerts.title}
                </Typography>
                <Typography className={this.classes.pos} color="textSecondary">
                    {PagesDoc.ongoingAlerts.sub_title}
                </Typography>
                <Typography component="p">
                    {PagesDoc.ongoingAlerts.short_desc}
                </Typography>
                {/* <Typography component="p">
                    <Badge showZero className={this.classes.badge} badgeContent={this.state.NbrActive} color="primary">
                        <Typography className={this.classes.padding}>Active</Typography>
                    </Badge>
                    <Badge showZero className={this.classes.badge} badgeContent={this.state.NbrSuppressed} color="primary">
                        <Typography className={this.classes.padding}>Suppressed</Typography>
                    </Badge>
                </Typography> */}
              </CardContent>
              <CardActions>
                <Button color="secondary" variant="contained" component={Link} to="/ongoing-alerts" size="small">Go</Button>
              </CardActions>
            </Card>
            <Card className={this.classes.card}>
              <CardContent>
                <Typography variant="h5" component="h2">
                {PagesDoc.alertsExplorer.title}
                </Typography>
                <Typography className={this.classes.pos} color="textSecondary">
                {PagesDoc.alertsExplorer.sub_title}
                </Typography>
                <Typography component="p">
                {PagesDoc.alertsExplorer.short_desc}
                </Typography>
              </CardContent>
              <CardActions>
                <Button color="secondary" variant="contained" component={Link} to="/alerts-explorer" size="small">Go</Button>
              </CardActions>
            </Card>
            <Card className={this.classes.card}>
              <CardContent>
                <Typography variant="h5" component="h2">
                {PagesDoc.suppressionRules.title}
                </Typography>
                <Typography className={this.classes.pos} color="textSecondary">
                {PagesDoc.suppressionRules.sub_title}
                </Typography>
                <Typography component="p">
                {PagesDoc.suppressionRules.short_desc}
                </Typography>
              </CardContent>
              <CardActions>
                <Button color="secondary" variant="contained" component={Link} to="/suppression-rules" size="small">Go</Button>
              </CardActions>
            </Card>
            </div>
            </div>
        );
    }
}


export default withStyles(styles)(HomeView);
