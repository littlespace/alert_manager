
import React from 'react';
import { withStyles } from '@material-ui/core/styles';


import Typography from '@material-ui/core/Typography';

import SuppressionRuleItem from '../components/SuppressionRule/SuppressionRuleItem'
import { AlertManagerApi } from '../library/AlertManagerApi';


import Grid from '@material-ui/core/Grid';

import PageHelp from '../components/PageHelp';
import { PagesDoc } from '../static';
import { SuppDurations } from './AlertView';
import Fab from '@material-ui/core/Fab';
import Tooltip from '@material-ui/core/Tooltip';
import Button from '@material-ui/core/Button';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogTitle from '@material-ui/core/DialogTitle';
import FormControl from '@material-ui/core/FormControl';
import TextField from '@material-ui/core/TextField';
import MenuItem from '@material-ui/core/MenuItem';
import Input from '@material-ui/core/Input';
import InputLabel from '@material-ui/core/InputLabel';
import FormHelperText from '@material-ui/core/FormHelperText';
import IconButton from '@material-ui/core/IconButton';
import Snackbar from '@material-ui/core/Snackbar';


/// -------------------------------------
/// Icons 
/// -------------------------------------
import HelpIcon from '@material-ui/icons/Help';
import AddIcon from '@material-ui/icons/Add';
import DeleteIcon from '@material-ui/icons/Delete';
import CloseIcon from '@material-ui/icons/Close';

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
    pageTitle: {
        height: "30px",
        lineHeight: "30px",
        paddingLeft: "15px",
        paddingTop: "10px"
    },
    titleBar: {
        display: "flex",
        position: "relative",
    },
    helpButton: {
        width: "30px",
        height: "30px",
        minHeight: "20px",
        marginTop: "9px",
        marginLeft: "10px",
    },
    textField: {
        width: 300,
    },
    formControl: {
        minWidth: 120,
    },
    form: {
        display: 'flex',
        flexDirection: 'column',
        margin: 'auto',
        width: 'fit-content',
    },
    divStyle: {
        display: 'flex',
        alignItems: 'center'
    },
});

const mconds = [
    {
        value: 1,
        label: "All",
    },
    {
        value: 0,
        label: "Any",
    },
]


class SuppressionRulesListView extends React.Component {

    constructor(props) {
        super(props);
        this.classes = this.props.classes;
        this.state = {
            rules: [],
            rule_create_dialog_open: false,
            snackbar_open: false,
            supprule: {
                duration: 3600,
                reason: "",
                name: "",
                mcond: 1,
                entities: {},
            },
            supprule_ents: [{ key: "", value: "" }],
        }
        this.openHelp = this.openHelp.bind(this)
        this.closeHelp = this.closeHelp.bind(this)
        this.snackbar_msg = ""
        this.api = new AlertManagerApi();
    };

    componentDidMount() {
        this.updateSuppRulesList()
    }

    updateSuppRulesList = async () => {
        const rules = await this.api.getSuppressionRuleDynamicList()
        const prules = await this.api.getSuppressionRulePersistentList()
        var allrules = rules.concat(prules)
        this.setState({ rules: allrules })
    }

    openHelp = () => {
        this.setState({ openHelp: true })
    }

    closeHelp = () => {
        this.setState({ openHelp: false })
    }

    handleRuleCreateDialogOpen = () => {
        this.setState({ rule_create_dialog_open: true });
    };

    handleRuleCreateDialogClose = () => {
        this.setState({ rule_create_dialog_open: false });
    }

    updateSuppRule = (event) => {
        const name = event.target.name;
        const value = event.target.value;

        this.setState({
            supprule: {
                ...this.state.supprule,
                [name]: value,
            }
        });
    }

    addSuppRuleEntity = () => {
        this.setState((prevState) => ({
            supprule_ents: [...prevState.supprule_ents, { key: "", value: "" }],
        }));
    }

    handleSuppRuleEntChange = (e, i, type) => {
        let ents = this.state.supprule_ents
        ents[i][type] = e.target.value
        this.setState({ supprule_ents: ents })
    }

    removeSuppRuleEnt = (i) => {
        let ents = this.state.supprule_ents
        ents.splice(i, 1)
        this.setState({ supprule_ents: ents })
    }

    createSuppRule = async (e) => {
        e.preventDefault()
        const req = this.state.supprule
        req['creator'] = this.api.getUsername()
        const ents = this.state.supprule_ents
        ents.forEach(function (ent) {
            req.entities[ent.key] = ent.value
        })
        this.handleRuleCreateDialogClose()
        const rule = await this.api.createSuppRule(req)
        this.snackbar_msg = `Created Suppression Rule ${rule.id}`
        this.updateSuppRulesList()
        this.handleSnackBarOpen()
    }

    async handleDelete(id) {
        await this.api.clearSuppRule({ id: id })
        this.snackbar_msg = `Cleared Suppression Rule ${id}`
        this.updateSuppRulesList()
        this.handleSnackBarOpen()
    }

    handleSnackBarOpen = () => {
        this.setState({ snackbar_open: true })
    }

    handleSnackBarClose = (e, reason) => {
        if (reason === 'clickaway') {
            return;
        }
        this.setState({ snackbar_open: false })
    }

    render() {
        let activeRules = this.state.rules.filter(
            (rule) => {
                if (rule.dont_expire) {
                    return true
                }
                const createdSecs = Date.parse(rule.created_at) / 1000
                return createdSecs + rule.duration > (new Date()).getTime() / 1000
            }
        )
        let ents = this.state.supprule_ents
        return (
            <div>
                <div className={this.classes.titleBar}>
                    <Typography className={this.classes.pageTitle} variant="h5">{PagesDoc.suppressionRules.title}</Typography>
                    <Tooltip title="Help">
                        <Fab
                            size="small"
                            color="primary"
                            aria-label="Help"
                            onClick={this.openHelp}
                            className={this.classes.helpButton}>
                            <HelpIcon />
                        </Fab>
                    </Tooltip>
                    <Tooltip title="Add" aria-label="Add">
                        <Fab
                            size="small"
                            color="primary"
                            aria-label="Help"
                            onClick={this.handleRuleCreateDialogOpen}
                            className={this.classes.helpButton}>
                            <AddIcon />
                        </Fab>
                    </Tooltip>
                    <PageHelp
                        title={PagesDoc.suppressionRules.title}
                        description={PagesDoc.suppressionRules.help}
                        open={this.state.openHelp}
                        close={this.closeHelp}
                        showSuppRuleLegent={true} />
                </div>
                {/* TODO : Move this dialog to its own component. */}
                <Dialog
                    open={this.state.rule_create_dialog_open}
                    fullWidth={true}
                    maxWidth={"md"}
                    aria-labelledby="rule-create"
                >
                    <DialogTitle id="rule-create-title">Create New Rule</DialogTitle>
                    <DialogContent>
                        <form className={this.classes.form} id="create-rule" onSubmit={this.createSuppRule}>
                            <TextField
                                select
                                className={this.classes.textField}
                                value={this.state.supprule.duration}
                                label="Duration"
                                name="duration"
                                onChange={this.updateSuppRule}
                                margin="normal"
                            >
                                {SuppDurations.map(option => (
                                    <MenuItem key={option.value} value={option.value}>
                                        {option.label}
                                    </MenuItem>
                                ))}
                            </TextField>
                            <TextField
                                id="standard-full-width"
                                className={this.classes.textField}
                                value={this.state.supprule.reason}
                                name="reason"
                                label="Suppress Reason"
                                onChange={this.updateSuppRule}
                                placeholder={"Suppressed by " + this.api.getUsername()}
                                fullWidth
                                margin="normal"
                                InputLabelProps={{
                                    shrink: true,
                                }}
                            />
                            <TextField
                                className={this.classes.textField}
                                value={this.state.supprule.name}
                                id="standard-required"
                                name="name"
                                label="Rule Name"
                                onChange={this.updateSuppRule}
                                margin="normal"
                                required={true}
                                InputLabelProps={{
                                    shrink: true,
                                }}
                            />
                            <TextField
                                select
                                className={this.classes.textField}
                                value={this.state.supprule.mcond}
                                name="mcond"
                                label="Match Condition"
                                onChange={this.updateSuppRule}
                                margin="normal"
                            >
                                {mconds.map(option => (
                                    <MenuItem key={option.value} value={option.value}>
                                        {option.label}
                                    </MenuItem>
                                ))}
                            </TextField>
                            {
                                ents.map((val, idx) => {
                                    let keyId = `key-${idx}`, valId = `val-${idx}`
                                    return (
                                        <div className={this.classes.divStyle}>
                                            <FormControl className={this.classes.formControl} required={true}>
                                                <InputLabel htmlFor={keyId}>Entity Key</InputLabel>
                                                <Input id={keyId} aria-describedby="key-helper-text" onChange={(e) => this.handleSuppRuleEntChange(e, idx, "key")}></Input>
                                                <FormHelperText id="key-helper-text">e.g device</FormHelperText>
                                            </FormControl>
                                            <FormControl className={this.classes.formControl} required={true}>
                                                <InputLabel htmlFor={valId}>Entity Value</InputLabel>
                                                <Input id={valId} aria-describedby="val-helper-text" onChange={(e) => this.handleSuppRuleEntChange(e, idx, "value")}></Input>
                                                <FormHelperText id="val-helper-text">Regexes accepted</FormHelperText>
                                            </FormControl>
                                            <IconButton disabled={idx === 0} onClick={() => this.removeSuppRuleEnt(idx)} aria-label="Delete">
                                                <DeleteIcon />
                                            </IconButton>
                                        </div>
                                    )
                                })
                            }
                            <div>
                                <IconButton onClick={this.addSuppRuleEntity} aria-label="Add">
                                    <AddIcon />
                                </IconButton>
                            </div>

                        </form>
                    </DialogContent>
                    <DialogActions>
                        <Button
                            color="default"
                            onClick={this.handleRuleCreateDialogClose}>
                            Close
                    </Button>
                        <Button
                            color="secondary"
                            type="submit"
                            form="create-rule"
                        >
                            Create
                    </Button>
                    </DialogActions>
                </Dialog>
                <div className={this.classes.root}>
                    <Grid
                        className={this.classes.gridroot}
                        container
                        spacing={16}
                        direction="column"
                        justify="flex-start"
                        alignItems="stretch"
                    >
                        {activeRules.map(r => {
                            return (
                                <SuppressionRuleItem
                                    key={r.id}
                                    data={r}
                                    onDelete={() => this.handleDelete(r.id)}
                                >
                                </SuppressionRuleItem >
                            );
                        })}
                    </Grid>
                </div>
                <div>
                    <Snackbar
                        anchorOrigin={{
                            vertical: 'bottom',
                            horizontal: 'left',
                        }}
                        open={this.state.snackbar_open}
                        autoHideDuration={5000}
                        onClose={this.handleSnackBarClose}
                        ContentProps={{
                            'aria-describedby': 'message-id',
                        }}
                        message={<span id="message-id">{this.snackbar_msg}</span>}
                        action={[
                            <IconButton
                                key="close"
                                aria-label="Close"
                                color="inherit"
                                onClick={this.handleSnackBarClose}
                            >
                                <CloseIcon />
                            </IconButton>,
                        ]}
                    />
                </div>

            </div>
        );
    }
}


export default withStyles(styles)(SuppressionRulesListView);
