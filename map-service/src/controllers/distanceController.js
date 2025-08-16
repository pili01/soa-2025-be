const axios = require('axios');
const API_KEY = process.env.MAP_API_KEY;

const VALID_PROFILES = ['driving-car', 'foot-walking', 'cycling-regular'];

const getDistance = async (req, res) => {
  try {
    const { originLat, originLng, destLat, destLng } = req.query;

    if (!originLat || !originLng || !destLat || !destLng) {
      return res.status(400).json({ error: 'Origin and destination coordinates are required' });
    }

    const results = {};

    for (const profile of VALID_PROFILES) {
      const url = `https://api.openrouteservice.org/v2/matrix/${profile}`;
      const response = await axios.post(
        url,
        {
          locations: [
            [parseFloat(originLng), parseFloat(originLat)],
            [parseFloat(destLng), parseFloat(destLat)]
          ],
          metrics: ['distance', 'duration']
        },
        {
          headers: {
            Authorization: API_KEY,
            'Content-Type': 'application/json'
          }
        }
      );

      const distance = response.data.distances[0][1]; // meters
      const duration = response.data.durations[0][1]; // seconds

      results[profile] = { distance, duration };
    }

    res.json(results);
  } catch (err) {
    console.error(err.response?.data || err.message);
    res.status(500).json({ error: 'Failed to calculate distance' });
  }
};

module.exports = { getDistance };