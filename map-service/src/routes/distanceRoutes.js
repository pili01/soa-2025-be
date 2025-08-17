const express = require('express');
const { getDistance } = require('../controllers/distanceController');

const router = express.Router();

router.get('/api/getdistances', getDistance);

module.exports = router;
