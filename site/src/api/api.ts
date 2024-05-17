interface ParkCam {
  title: string;
  image: string;
  link: string;
}

const BACKEND_URL = 'http://localhost:4001';
export function FetchParkCams(): Promise<ParkCam[]> {
  return fetch(`${BACKEND_URL}/park-cams`)
    .then(response => {
      if (!response.ok) {
        throw new Error('Network response was not ok');
      }
      return response.json();
    })
    .then(data => {
      return data;
    });
}

export function FetchParkList(): Promise<string[]> {
  return fetch(`${BACKEND_URL}/park-list`)
    .then(response => {
      if (!response.ok) {
        throw new Error('Network response was not ok');
      }
      return response.json();
    })
    .then(data => {
      return data.parks;
    });
}