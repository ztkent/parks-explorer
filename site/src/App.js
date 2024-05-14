import * as React from 'react';
import * as Components from './components';
import { createTheme } from '@mui/material/styles';
import { CssBaseline, ThemeProvider, Toolbar } from '@mui/material';

const theme = createTheme({
  palette: {
    mode: 'dark',
  },
});

function App() {
  const [open, setOpen] = React.useState(false);

  const handleDrawerOpen = () => {
    setOpen(true);
  };

  const handleDrawerClose = () => {
    setOpen(false);
  };

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Components.Header handleDrawerOpen={handleDrawerOpen} />
      <Toolbar /> {/* This acts as a spacer to prevent the content from being hidden under the AppBar */}
      <Components.Sidebar open={open} handleDrawerClose={handleDrawerClose} />
      <Components.MainContent
                display="flex"
                justifyContent="center"
                alignItems="center"
        />
    </ThemeProvider>
  );
}

export default App;