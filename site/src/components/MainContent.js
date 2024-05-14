import * as React from 'react';
import { Container, Box, Typography } from '@mui/material';

function MainContent() {
  return (
    <Container>
      <Box my={2} textAlign="center">
        <Typography variant="h4" component="h1" gutterBottom>
          Welcome to My Website
        </Typography>
        <Typography variant="body1">This is a simple landing page built with MUI.</Typography>
      </Box>
    </Container>
  );
}

export default MainContent;
