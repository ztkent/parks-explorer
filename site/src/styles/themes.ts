import { ThemeOptions } from '@mui/system';

export const naturalHarmonyTheme = (darkMode: boolean): ThemeOptions => ({
  palette: {
      mode: darkMode ? 'dark' : 'light',
      primary: {
        main: darkMode ? '#276738' : '#4CAF50',
        light: darkMode ? '#276738' : '#4CAF50',
        dark:  darkMode ? '#276738' : '#4CAF50',
        contrastText: darkMode ? '#fff' : '#000',
      },
      secondary: {
        main: darkMode ? '#276738' : '#4CAF50',
        light: darkMode ? '#276738' : '#4CAF50',
        dark:  darkMode ? '#276738' : '#4CAF50',
        contrastText: darkMode ? '#fff' : '#000',
      },
      error: {
        main: '#F44336',
        contrastText: '#fff',
      },
      warning: {
        main: '#FFA726',
        contrastText: '#000',
      },
      info: {
        main: '#BBDEFB',
        contrastText: '#000',
      },
      success: {
        main: '#66BB6A',
        contrastText: '#000',
      },
      background: {
        default: darkMode ? '#37474F' : '#E8F5E9',
        paper: darkMode ? '#455A64' : '#F1F8E9',
      },
      text: {
        primary: darkMode ? '#fff' : '#000',
        secondary: darkMode ? '#B0BEC5' : '#757575',
      },
    },
  });