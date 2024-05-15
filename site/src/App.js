import React, { useState } from 'react';
import * as Components from './components';
import { createTheme, ThemeProvider, CssBaseline, Switch, FormControlLabel} from '@mui/material';
import Brightness7Icon from '@mui/icons-material/Brightness7';
import Brightness4Icon from '@mui/icons-material/Brightness4';
import Toolbar from '@mui/material/Toolbar';
import { naturalHarmonyTheme} from './styles/themes.ts';
function App() {
  // Handle the theme selection
  const [darkMode, setDarkMode] = useState(true);
  const [themeOptions, setThemeOptions] = useState(naturalHarmonyTheme(darkMode));
  const handleThemeChange = () => {
    setDarkMode(!darkMode);
    setThemeOptions(naturalHarmonyTheme(!darkMode));
  };
  const theme = createTheme(themeOptions);

  // Handle the sidebar open and close state
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
      <Toolbar />
      <Components.Sidebar open={open} handleDrawerClose={handleDrawerClose} />
      <Components.MainContent
        display="flex"
        justifyContent="center"
        alignItems="center"
      />
      <FormControlLabel
        control={<Switch checked={darkMode} onChange={handleThemeChange} />}
        label={darkMode ? <Brightness4Icon /> : <Brightness7Icon />}
        labelPlacement="start"
        style={{ position: 'fixed', bottom: 15, right: 15 }}
      />
    </ThemeProvider>
  );
}

export default App;