const stepsDict = {
  0: "bytes",
  1: "kB", // kilobyte
  2: "MB", // megabyte
  3: "GB", // gigabyte
  4: "TB", // terabyte
  5: "PB", // petabyte
  6: "EB", // exabyte
  7: "ZB", // zetabyte
  8: "YB", // yotabyte
};

const sizesDict = {
  "bytes": 0,
  "kB": 1,
  "MB": 2,
  "GB": 3,
  "TB": 4,
  "PB": 5,
  "EB": 6,
  "ZB": 7,
  "YB": 8,
};

export const fetchData = async (url, options) => {
  // options.headers["Access-Control-Allow-Origin"] = "*";
  const response = await fetch("/api" + url, options);
  // if (!response.ok) {
  //     throw new Error(`Error fetching data. Server replied ${response.status}`);
  // }
  return response;
};

export const stepsToString = (steps) => (stepsDict)[steps];
export const exponentSwitch = (label) => (sizesDict)[label];

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

export const getSizeOptions = (storageSizeLabel) => {
  var options = [];
  var len = exponentSwitch(storageSizeLabel);
  for(var i = 0; i <= len; i++) {
    options.push(stepsDict[i]);
  }
  return options;
}

// export const getClient = (loadingHandler, id, token, setClientToken) => {
//   loadingHandler("get_client", "push");
//   const url = "/clients/" + id;
//   const options = {
//     method: 'GET',
//     headers: {
//         'Authorization': 'Basic ' + token,
//         'Content-Type': 'application/x-www-form-urlencoded',
//     },
//     body: "name=" + id,
//   };
//   fetchData(url, options)
//   .then(result => result.json())
//   .then(result => {
//     setClientToken(result.AuthCode);
//     loadingHandler("get_client", "pop");
//   }, err => {
//     console.log(err);
//     loadingHandler("get_client", "pop");
//   });
// };

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

export const fetchClients = async (token) => {
  const url = "/clients";
  const options = {
    headers: {
      'Authorization': 'Basic ' + token,
    },
  };
  return fetchData(url, options).then((response) => {
    if(response.ok) {
      return response.json();
    } else {
      return null;
    }  
  });
};

export const fetchStorageSpace = async (token) => {
  const url = "/storage_size";
  const options = {
    headers: {
      'Authorization': 'Basic ' + token,
    },
  };
  const response = await fetchData(url, options);
  return await response.json();
};

export const fetchStorageSpaceMinusQuota = async (token) => {
  const url = "/storage_size_minus_quota";
  const options = {
    headers: {
      'Authorization': 'Basic ' + token,
    },
  };
  const response = await fetchData(url, options);
  return await response.json();
};

export const fetchStorageSizePlusQuota = async (token, client) => {
  const url = "/storage_size_plus_quota?id=" + client.ID;
  const options = {
    headers: {
      'Authorization': 'Basic ' + token,
    },
  };
  const response = await fetchData(url, options);
  return await response.json();
};

export const fetchUsedSpace = async (token) => {
  const url = "/used_space";
  const options = {
    headers: {
      'Authorization': 'Basic ' + token,
    },
  };
  const response = await fetchData(url, options);
  return await response.json();
};

export const fetchTotalQuota = async (token) => {
  const url = "/total_quota";
  const options = {
    headers: {
      'Authorization': 'Basic ' + token,
    },
  };
  const response = await fetchData(url, options);
  return await response.json();
};

export const fetchServerInformation = async (token) => {
  const url = "/server_config";
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
