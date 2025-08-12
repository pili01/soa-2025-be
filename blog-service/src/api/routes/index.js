const express = require('express');
const blogRoutes = require('./blogRoutes');

const router = express.Router();

router.use('/blogs', blogRoutes);

module.exports = router;