const likeService = require('../../service/likeService');
const authService = require('../../service/authService');
const followerService = require('../../service/followerService');
const blogService = require('../../service/blogService');

exports.create = async (req, res, next) => {
  try {
    const data = await authService.getMe(req.headers.authorization);
    if (!data || !data.role || !data.userId) {
      return res.status(403).json({
        success: false,
        message: "Forbidden"
      });
    }
    const likeData = req.body;
    const blogId = likeData.blogId;
    if (!blogId) {
      return res.status(400).json({
        success: false,
        message: "Blog ID is required"
      });
    }
    const blog = await blogService.getBlogById(blogId);
    if (!blog) {
      return res.status(404).json({
        success: false,
        message: "Blog not found"
      });
    }
    console.log("Checking if user is followed...");
    if (!(await followerService.isUserFollowedByMe(req.headers.authorization, blog.userId))) {
      return res.status(403).json({
        success: false,
        message: "Forbidden: You are not allowed to like this blog, you must follow the author."
      });
    }
    likeData.userId = data.userId;

    const newLike = await likeService.createLike(likeData);

    res.status(201).json({
      success: true,
      message: 'You managed to create new like on blog',
      data: newLike
    });
  } catch (err) {
    res.status(500).json({
      success: false,
      message: 'An error occurred on the server.',
    });
  }
}

exports.getAllBlogLikes = async (req, res, next) => {
  try {
    const { blogId } = req.params;
    if (!blogId) {
      return res.status(400).json({ success: false, message: 'Missing blogId in params.' });
    }

    const likes = await likeService.getAllBlogLikes(+blogId);

    return res.status(200).json({
      success: true,
      message: 'Fetched all likes for the blog.',
      data: likes
    });
  } catch (err) {
    return res.status(500).json({
      success: false,
      message: 'An error occurred on the server.'
    });
  }
};

exports.deleteBlogLike = async (req, res, next) => {
  try {
    const { blogId, userId } = req.params;
    if (!blogId || !userId) {
      return res.status(400).json({ success: false, message: 'Missing blogId or userId in params.' });
    }

    const deleted = await likeService.delete(+blogId, +userId);
    if (!deleted) {
      return res.status(404).json({ success: false, message: 'Like not found for the given blogId/userId.' });
    }

    return res.status(204).send();
  } catch (err) {
    return res.status(500).json({ success: false, message: 'An error occurred on the server.' });
  }
};