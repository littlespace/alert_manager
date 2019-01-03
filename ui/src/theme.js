import { createMuiTheme } from 'material-ui/styles';
import red from 'material-ui/colors/red';

export default createMuiTheme({
    palette: {
      primary: {
        main: '#424242',
      },
      secondary: red,
    },
    status: {
      danger: 'orange',
    },
  });
