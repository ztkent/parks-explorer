import * as React from "react";
import {Drawer, List, ListItemButton, ListItemText, Toolbar} from "@mui/material";

function Sidebar({ open, handleDrawerClose }) {
  return (
    <Drawer
      variant="temporary"
      open={open}
      onClose={handleDrawerClose}
      ModalProps={{
        // Better open performance on mobile.
        keepMounted: true,
      }}
    >
      <Toolbar /> {/* This acts as a spacer to prevent the content from being hidden under the AppBar */}
      <List
        sx={{
          // Set the sidebar width to 240px
          minWidth: "240px",
        }}
      >
        {["Item 1", "Item 2", "Item 3"].map((text, index) => (
          <ListItemButton key={text}>
            <ListItemText primary={text} sx={{ textAlign: 'center' }} />
          </ListItemButton>
        ))}
      </List>
    </Drawer>
  );
}

export default Sidebar;