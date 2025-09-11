const express = require('express');
const likeController = require('../controllers/likeController');

const router = express.Router();

router.post('/:blogId/likes', likeController.create);
router.get('/:blogId/likes', likeController.getAllBlogLikes);
router.delete('/:blogId/likes/:userId', likeController.deleteBlogLike);
router.get('/:blogId/likes/me', likeController.getMyLikeStatus);
router.post('/:blogId/likes/toggle', likeController.toggleLike);
router.get('/:blogId/likes/count', likeController.getBlogLikesCount);

module.exports = router;