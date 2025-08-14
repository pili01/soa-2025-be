const express = require('express');
const blogRoutes = require('./blogRoutes');
const likeRoutes = require('./likeRoutes');
const commentRoutes = require('./commentRoutes');

const router = express.Router();

router.use('/blogs', blogRoutes);
router.use('/blogs', likeRoutes);
router.use('/blogs', commentRoutes);

module.exports = router;