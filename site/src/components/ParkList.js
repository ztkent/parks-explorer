import React from 'react';
import { Grid, Box, Typography, Button, ButtonGroup, useTheme } from '@mui/material';

function ParkList({ parkListData, visibleEntries, loadMoreEntries, handleSortChange, sortOption }) {
  const sortOptions = ['Alphabetical', 'Most Popular', 'Recently Added'];
  const theme = useTheme();

  return (
    <Box my={4} textAlign="center">
      <Typography variant="h5" component="h2" gutterBottom>
        All of our Parks
      </Typography>

      <Box mb={2}>
        <Typography variant="body1" component="p">
          Sort by:
        </Typography>
        <ButtonGroup
            aria-label="outlined secondary button group"
            sx={{
                '& .MuiButton-root': {
                color: theme.palette.mode === 'dark' ? 'white' : 'black',
                border: 1,
                '&:hover': {
                    background: theme.palette.primary.main,
                  },
                },
            }}
            >
            {sortOptions.map((option, idx) => (
                <Box marginRight={0} key={idx}> {}
                <Button onClick={() => handleSortChange(option)}>
                    {option}
                </Button>
                </Box>
            ))}
        </ButtonGroup>
      </Box>

      <Grid container spacing={3}>
        {[0, 1, 2].map((col) => (
          <Grid item xs={12} sm={4} key={col}>
            <Box>
              {parkListData.slice(col * visibleEntries/3, (col + 1) * visibleEntries/3).map((entry, idx) => (
                <Typography variant="body2" key={idx}>
                  {entry}
                </Typography>
              ))}
            </Box>
          </Grid>
        ))}
      </Grid>
      {visibleEntries < parkListData.length && (
        <Button variant="contained" color="secondary" onClick={loadMoreEntries} style={{ marginTop: '16px' }}>
          Load More
        </Button>
      )}
    </Box>
  );
}

export default ParkList;