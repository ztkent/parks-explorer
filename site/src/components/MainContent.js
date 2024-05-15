import * as React from 'react';
import { 
  Container, Divider 
} from '@mui/material';
import SearchBox from './SearchBox';
import LiveParks from './LiveParks';
import ParkList from './ParkList';

// Dummy Data
function fetchData() {
  return new Promise((resolve) => {
    setTimeout(() => {
      const displayComponentsData = Array.from({ length: 9 }, (_, idx) => ({
        title: `Item ${idx + 1}`,
        image: 'https://via.placeholder.com/150',
        link: '#',
      }));
      const columnData = Array.from({ length: 400 }, (_, idx) => `Entry ${idx + 1}`);

      resolve({ displayComponentsData, columnData });
    }, 2000); // Simulate 2 seconds network latency
  });
}

function MainContent() {
  // [state variable, handler function] = React.useState(initialValue)
  // State for the dropdown menu anchor element
  const [anchorEl, setAnchorEl] = React.useState(null);
  // State for the placeholder text in the search box
  const [placeholder, setPlaceholder] = React.useState("Search parks...");
  // State for the selected sort option for the 'All of our Parks' section
  const [sortOption, setSortOption] = React.useState(null);
  // State for the number of visible entries in the 'All of our Parks' section
  const [visibleEntries, setVisibleEntries] = React.useState(30);
  // State for the display components data
  const [displayComponentsData, setDisplayComponentsData] = React.useState([]);
  // State for the column data
  const [columnData, setColumnData] = React.useState([]);

  React.useEffect(() => {
    fetchData().then(({ displayComponentsData, columnData }) => {
      setDisplayComponentsData(displayComponentsData);
      setColumnData(columnData);
    });
  }, []);

  const handleDropdownClick = (event) => {
    setAnchorEl(event.currentTarget);
  };

  const handleDropdownClose = () => {
    setAnchorEl(null);
  };

  const handleSelectOption = (option) => {
    setPlaceholder(option);
    setSortOption(option);
    setAnchorEl(null);
  };

  const loadMoreEntries = () => {
    setVisibleEntries((prevVisibleEntries) => prevVisibleEntries + 30);
  };

  const handleSortChange = (option) => {
    setSortOption(option);
    sortColumnData(option);
  };

  const sortColumnData = (option) => {
    let sortedData = [...columnData];
    switch (option) {
      case 'Alphabetical':
        sortedData.sort();
        break;
      case 'Most Popular':
      case 'Recently Added':
        // Randomly sort the data
        sortedData.sort(() => Math.random() - 0.5);
        break;
      default:
        break;
    }
    setColumnData(sortedData);
  };

  return (
    <Container>
      <SearchBox 
        placeholder={placeholder}
        handleDropdownClick={handleDropdownClick}
        handleDropdownClose={handleDropdownClose}
        handleSelectOption={handleSelectOption}
        anchorEl={anchorEl}
      />

      <Divider />

      <LiveParks displayComponentsData={displayComponentsData} />

      <Divider />

      <ParkList 
        columnData={columnData}
        visibleEntries={visibleEntries}
        loadMoreEntries={loadMoreEntries}
        handleSortChange={handleSortChange}
        sortOption={sortOption}
      />
    </Container>
  );
}

export default MainContent;