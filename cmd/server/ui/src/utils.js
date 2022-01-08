export const fetchData = async (url, options) => {
    const response = await fetch(url, options);
    if (!response.ok) {
        throw new Error(`Error fetching data. Server replied ${response.status}`);
    }
    console.log(response);
    return response;
};