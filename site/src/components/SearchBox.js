import React, { useState } from 'react';
import { TextField, Button, IconButton, Menu, MenuItem, Box, Typography, Chip, InputAdornment } from '@mui/material';
import ArrowDropDownIcon from '@mui/icons-material/ArrowDropDown';

function SearchBox({ placeholder, handleDropdownClick, handleDropdownClose, handleSelectOption, anchorEl, onSearch }) {
  const [inputValue, setInputValue] = useState('');
  const [inputValues, setInputValues] = useState([]);

  const handleInputChange = (event) => {
    setInputValue(event.target.value);
  };

  const handleInputKeyDown = (event) => {
    if (event.key === 'Enter') {
      event.preventDefault();
      if (inputValue && !inputValues.includes(inputValue)) {
        setInputValues([...inputValues, inputValue]);
        setInputValue('');
      }
    }
  };

  const handleDelete = () => {
    setInputValues(inputValues.slice(0, -1));
  };

  const handleSearch = () => {
    onSearch(inputValues);
    setInputValues([]);
  };

  return (
    <Box my={8} textAlign="center">
      <Typography variant="h4" component="h2" gutterBottom>
        Find Your Park
      </Typography>
      <Box display="flex" justifyContent="center" alignItems="center">
        <TextField 
          value={inputValue}
          onChange={handleInputChange}
          onKeyDown={handleInputKeyDown}
          variant="outlined" 
          style={{ marginRight: '8px', width: '75%' }} 
          placeholder={placeholder}
          InputProps={{
            startAdornment: inputValues.length >= 4 ? (
              <InputAdornment position="start">
                <Chip
                  label={`+${inputValues.length} tags`}
                  onDelete={handleDelete}
                  style={{ marginRight: '8px' }}
                />
              </InputAdornment>
            ) : (
              inputValues.map((item, index) => (
                <InputAdornment position="start" key={index}>
                  <Chip
                    label={item}
                    onDelete={() => handleDelete(item)}
                    style={{ marginRight: '8px' }}
                  />
                </InputAdornment>
              ))
            )
          }}
        />
        <Button 
          variant="contained" 
          color="primary" 
          style={{ height: '56px' }} 
          onClick={handleSearch}
        >
          Search
        </Button>
        <IconButton onClick={handleDropdownClick}>
          <ArrowDropDownIcon />
        </IconButton>
        <Menu
          anchorEl={anchorEl}
          open={Boolean(anchorEl)}
          onClose={handleDropdownClose}
        >
          <MenuItem onClick={() => handleSelectOption("Search parks...")}>Search parks...</MenuItem>
          <MenuItem onClick={() => handleSelectOption("Search activities...")}>Search activities...</MenuItem>
          <MenuItem onClick={() => handleSelectOption("Search events...")}>Search events...</MenuItem>
        </Menu>
      </Box>
    </Box>
  );
}

export default SearchBox;