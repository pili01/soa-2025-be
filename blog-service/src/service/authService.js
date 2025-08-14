const axios = require('axios');

exports.getMe = async (authHeader) => {
    try {
        const response = (await axios.get('http://stakeholders-service:8080/api/me', {
            headers: {
                Authorization: authHeader
            }
        }));
        return response.data;
    } catch (error) {
        console.error('Error fetching user data:', error);
        throw error;
    }
}
