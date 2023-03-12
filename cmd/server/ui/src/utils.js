export const fetchData = async (url, options) => {
    const response = await fetch(url, options);
    // if (!response.ok) {
    //     throw new Error(`Error fetching data. Server replied ${response.status}`);
    // }
    return response;
};

export const stepsToString = (steps) => {
  const stepSwitch = (step) => ({
    0: "bytes",
    1: "kB", // kilobyte
    2: "MB", // megabyte
    3: "GB", // gigabyte
    4: "TB", // terabyte
    5: "PB", // petabyte
    6: "EB", // exabyte
    7: "ZB", // zetabyte
    8: "YB", // yotabyte
  })[step];

  return stepSwitch(steps);
};

export const exponentSwitch = (label) => ({
  "bytes": 0,
  "kB": 1,
  "MB": 2,
  "GB": 3,
  "TB": 4,
  "PB": 5,
  "EB": 6,
  "ZB": 7,
  "YB": 8,
})[label];

export const sizeToBytes = (size, sizeLabel) => {
  return size * Math.pow(1000, exponentSwitch(sizeLabel));
};

export const convertSizeByStep = (size, steps) => {
  var resultSize = size;
  for(var i = 0; i < steps; i++) {
    resultSize = Math.floor(resultSize/1000);
  }
  return resultSize;
}

export const sizeConversion = (size, steps) => {
  if(size < 10000) {
    return [size, stepsToString(steps)];
  }

  return sizeConversion(Math.floor(size/1000), ++steps);
};

export const getClient = (setIsLoading, id, token, setClientToken) => {
  setIsLoading(true);
  const url = "/clients/" + id;
  const options = {
    method: 'GET',
    headers: {
        'Authorization': 'Basic ' + token,
        'Content-Type': 'application/x-www-form-urlencoded',
    },
    body: "name=" + id,
  };
  fetchData(url, options)
  .then(result => result.json())
  .then(result => {
    setClientToken(result.AuthCode);
    setIsLoading(false);
  }, err => {
    console.log(err);
    setIsLoading(false);
  });
};

export const fetchClient = async (token, id) => {
  const url = "/clients/" + id;
  const options = {
    headers: {
      'Authorization': 'Basic ' + token,
    },
  };
  const response = await fetchData(url, options);
  return await response.json();
};

export const fetchClients = async (navigate, token) => {
  const url = "/clients";
  const options = {
    headers: {
      'Authorization': 'Basic ' + token,
    },
  };
  const response = await fetchData(url, options);
  if(response.ok) {
    return await response.json();
  } else {
    navigate("/login");
    return null;
  }
};

export const fetchStorageSize = async (token) => {
  const url = "/storage_size";
  const options = {
    headers: {
      'Authorization': 'Basic ' + token,
    },
  };
  const response = await fetchData(url, options);
  return await response.json();
};

export const createClientRequest = async (token, client) => {
  const fetchUrl = "/clients";
  const fetchOptions = {
    method: 'POST',
    headers: {
        'Authorization': 'Basic ' + token,
        'Content-Type': 'application/x-www-form-urlencoded',
    },
    body: "name=" + client.Name + "&quota=" + client.Quota,
  };
  return await fetchData(fetchUrl, fetchOptions);
};

export const updateClientRequest = async (token, client) => {
  const url = "/clients/" + client.ID;
  const options = {
    method: 'PUT',
    headers: {
      'Authorization': 'Basic ' + token,
      'Content-Type': 'application/x-www-form-urlencoded',
    },
    body: "name=" + client.Name + "&quota=" + client.Quota,
  };
  return await fetchData(url, options);
};

export const deleteClientRequest = async (token, client_id) => {
  const fetchUrl = "/clients/" + client_id;
  const fetchOptions = {
    method: 'DELETE',
    headers: {
        'Authorization': 'Basic ' + token,
    },
  }
  return await fetchData(fetchUrl, fetchOptions);
};
