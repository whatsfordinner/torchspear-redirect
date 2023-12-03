function handler(event) {
    var response = {
        statusCode: 301,
        statusDescription: 'Found',
        headers: {
            'location': { value: 'https://www.oglaf.com/endpoint/' }
        }
    };
    return response;
}