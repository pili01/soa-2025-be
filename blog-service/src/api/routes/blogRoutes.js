const express = require('express');
const blogController = require('../controllers/blogController');

const router = express.Router();

router.post('/', blogController.createBlog);
router.get('/', blogController.getBlogs);
router.get('/all', blogController.getAllBlogs);
router.get('/:id', blogController.getBlogById);

module.exports = router;