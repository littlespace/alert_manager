
import React from 'react';
import { withStyles } from '@material-ui/core/styles';


import Typography from '@material-ui/core/Typography';

import SuppressionRuleItem from '../components/SuppressionRule/SuppressionRuleItem'

import { AlertManagerApi } from '../library/AlertManagerApi';

import Grid from '@material-ui/core/Grid';


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

    render() {
        let rules = this.state.rules
        return (
            <div>
                <Typography className={this.classes.pageTitle} variant="headline">Suppression Rules</Typography>   
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
                        </SuppressionRuleItem>
                        );
                    })}
                    </Grid>
                </div>
            </div>
        );
    }
}


export default withStyles(styles)(SuppressionRulesListView);
