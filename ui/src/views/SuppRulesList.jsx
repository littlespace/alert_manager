
import React from 'react';
import { withStyles } from '@material-ui/core/styles';


import Typography from '@material-ui/core/Typography';

import SuppressionRuleItem from '../components/SuppressionRule/SuppressionRuleItem'
import { AlertManagerApi } from '../library/AlertManagerApi';

import Grid from '@material-ui/core/Grid';

import PageHelp from '../components/PageHelp';
import { PagesDoc } from '../static';
import Fab from '@material-ui/core/Fab';
import Tooltip from '@material-ui/core/Tooltip';

/// -------------------------------------
/// Icons 
/// -------------------------------------
import HelpIcon from '@material-ui/icons/Help';

const AM = new AlertManagerApi()

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
    titleBar:{
        display: "flex",
        position: "relative",
      }, 
    helpButton: {
        width:  "30px",
        height: "30px",
        minHeight:  "20px",
        marginTop: "9px",
        marginLeft: "10px",
      }
});

class SuppressionRulesListView extends React.Component {

    constructor(props){
        super(props);
        this.classes = this.props.classes;
        this.state = {
            rules: [],
            rules_dynamic: [],
            rules_persistent: []
        }
        this.openHelp = this.openHelp.bind(this)
        this.closeHelp = this.closeHelp.bind(this)
    };

    componentDidMount(){
        this.updateSuppRulesList()
    }

    updateSuppRulesList = () => {
        AM.getSuppressionRulePersistentList()
          .then(data => this.updateSuppRulesPersistentList(data));

        AM.getSuppressionRuleDynamicList()
          .then(data => this.updateSuppRulesDynamicList(data));
    }

    updateSuppRulesPersistentList(rules) {

        let global_rules = rules.concat(this.state.rules_dynamic)

        this.setState({ rules_persistent: rules })
        this.setState({ rules: global_rules })

    }

    updateSuppRulesDynamicList(rules) {

        let global_rules = rules.concat(this.state.rules_persistent)

        this.setState({ rules_dynamic: rules })
        this.setState({ rules: global_rules })

    }

    openHelp = () => { 
        this.setState({ openHelp: true })
    }

    closeHelp = () => { 
        this.setState({ openHelp: false })
    }

    render() {
        let rules = this.state.rules
        return (
            <div>
                <div className={this.classes.titleBar}>
                <Typography className={this.classes.pageTitle} variant="h5">{PagesDoc.suppressionRules.title}</Typography>   
                <Tooltip title="help">
                    <Fab 
                        size="small" 
                        color="primary" 
                        aria-label="help" 
                        onClick={this.openHelp}
                        className={this.classes.helpButton}>
                        <HelpIcon />
                    </Fab>
                </Tooltip>
                <PageHelp 
                    title={PagesDoc.suppressionRules.title} 
                    description={PagesDoc.suppressionRules.help} 
                    open={this.state.openHelp}
                    close={this.closeHelp}
                    showSuppRuleLegent={true} />
                </div> 
                <div className={this.classes.root}>
                    <Grid 
                        className={this.classes.gridroot} 
                        container
                        spacing={16}
                        direction="column"
                        justify="flex-start"
                        alignItems="stretch"
                    >
                    { rules.map(r => {
                        return (
                        <SuppressionRuleItem
                            key={r.Id}
                            data={r}
                        >
                        </SuppressionRuleItem >
                        );
                    })}
                    </Grid>
                </div>
            </div>
        );
    }
}


export default withStyles(styles)(SuppressionRulesListView);