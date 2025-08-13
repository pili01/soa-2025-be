const express = require('express');
const likeController = require('../controllers/likeController');

const router = express.Router();

router.post('/likes', likeController.create);
router.get('/:blogId/likes', likeController.getAllBlogLikes);
router.delete('/:blogId/likes/:userId', likeController.deleteBlogLike);

module.exports = router;