import * as React from 'react';
import { 
  Container, Divider 
} from '@mui/material';
import SearchBox from './SearchBox';
import LiveParks from './LiveParks';
import ParkList from './ParkList';
import { FetchParkCams, FetchParkList } from '../api/api.ts';

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
  // State for the park cams data
  const [parkListData, setParkListData] = React.useState([]);
  // State for the park list data
  const [liveParkCamsData, setLiveParkCamsData] = React.useState([]);

  React.useEffect(() => {
    FetchParkList().then(({ parkListData }) => {
      setParkListData(parkListData);
    });
    FetchParkCams().then(({ liveParkCams }) => {
      setLiveParkCamsData(liveParkCams);
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
    sortParkListData(option);
  };

  const onSearch = (inputValues) => {
    console.log('Search:', inputValues);
  }

  const sortParkListData = (option) => {
    let sortedData = [...parkListData];
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
    setParkListData(sortedData);
  };

  return (
    <Container>
      <SearchBox 
        placeholder={placeholder}
        handleDropdownClick={handleDropdownClick}
        handleDropdownClose={handleDropdownClose}
        handleSelectOption={handleSelectOption}
        anchorEl={anchorEl}
        onSearch={onSearch}
      />

      <Divider />

      <LiveParks liveParkCamsData={liveParkCamsData} />

      <Divider />

      <ParkList 
        parkListData={parkListData}
        visibleEntries={visibleEntries}
        loadMoreEntries={loadMoreEntries}
        handleSortChange={handleSortChange}
        sortOption={sortOption}
      />
    </Container>
  );
}

export default MainContent;