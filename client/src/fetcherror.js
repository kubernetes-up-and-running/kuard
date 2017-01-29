export default function(response) {
    if (!response.ok) {
        throw Error(response.statusText);
    }
    return response;
}
