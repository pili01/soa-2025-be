const axios = require('axios');

exports.getFollowedUsers = async (authHeader) => {
    try{
        return (await axios.get('http://follower-service:8080/api/follow/followedByMe',{
            headers: {
                Authorization: authHeader
            }
        })).data;
    }catch (error) {
        console.error("Error fetching followed users:", error);
        throw error;
    }
}

exports.isUserFollowedByMe = async (authHeader, userId) => {
    try{
        return (await axios.get(`http://follower-service:8080/api/follow/followedByMe/${userId}`,{
            headers: {
                Authorization: authHeader
            }
        })).data.value || false;
    }catch (error) {
        console.error("Error fetching followed users:", error);
        throw error;
    }
}