const express = require('express');
const commentController = require('../controllers/commentController');

const router = express.Router();

router.post('/comment', commentController.createComment);
router.get('/comment/:blogId', commentController.getCommentsByBlogId);

module.exports = router;