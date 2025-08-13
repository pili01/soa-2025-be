const express = require('express');
const blogRoutes = require('./blogRoutes');
const likeRoutes = require('./likeRoutes');

const router = express.Router();

router.use('/blogs', blogRoutes);
router.use('/blogs', likeRoutes);

module.exports = router;