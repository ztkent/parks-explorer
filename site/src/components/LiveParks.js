import React from 'react';
import { Grid, Box, Typography, Button } from '@mui/material';

function LiveParks({ liveParkCamsData }) {
  return (
    <Box my={4} py={4}>
      <Typography variant="h5" component="h2" gutterBottom textAlign="center">
        Live at our Parks
      </Typography>
      <Grid container spacing={3}>
        {liveParkCamsData.map((item, index) => (
          <Grid item xs={12} sm={6} md={4} key={index}>
            <Box textAlign="center">
              <img src={item.image} alt={item.title} style={{ width: '100%' }} />
              <Typography variant="h6" component="h3">
                {item.title}
              </Typography>
              <Button variant="contained" color="primary" href={item.link} target="_blank">
                Learn More
              </Button>
            </Box>
          </Grid>
        ))}
      </Grid>
    </Box>
  );
}

export default LiveParks;