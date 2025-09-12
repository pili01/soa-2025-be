const express = require('express');
const commentController = require('../controllers/commentController');

const router = express.Router();

router.post('/comment', commentController.createComment);
router.get('/comment/:blogId', commentController.getCommentsByBlogId);
router.put('/comment/:commentId', commentController.updateComment);
router.delete('/comment/:commentId', commentController.deleteComment);

module.exports = router;