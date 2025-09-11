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

    const { blogId } = req.params;
    if (!blogId) {
      return res.status(400).json({
        success: false,
        message: "Blog ID is required"
      });
    }

    const blog = await blogService.getBlogById(+blogId);
    if (!blog) {
      return res.status(404).json({
        success: false,
        message: "Blog not found"
      });
    }

    const already = await likeService.hasUserLiked(Number(data.userId), Number(blogId));
    if (already) {
      return res.status(409).json({ success: false, message: "Already liked" });
    }

    if (!(await followerService.isUserFollowedByMe(req.headers.authorization, blog.userId))) {
      return res.status(403).json({
        success: false,
        message: "Forbidden: You must follow the author before liking."
      });
    }

    const likeData = { blogId: Number(blogId), userId: Number(data.userId) };
    const newLike = await likeService.createLike(likeData);

    res.status(201).json({
      success: true,
      message: 'You liked the blog.',
      data: newLike
    });
  } catch (err) {
    console.error(err);
    res.status(500).json({
      success: false,
      message: 'An error occurred on the server.'
    });
  }
};

exports.getAllBlogLikes = async (req, res, next) => {
  try {
    const { blogId } = req.params;
    if (!blogId) {
      return res.status(400).json({ success: false, message: 'Missing blogId in params.' });
    }

    const likes = await likeService.getAllBlogLikes(Number(blogId));

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

    const deleted = await likeService.delete(Number(blogId), Number(userId));
    if (!deleted) {
      return res.status(404).json({ success: false, message: 'Like not found for the given blogId/userId.' });
    }

    return res.status(204).send();
  } catch (err) {
    return res.status(500).json({ success: false, message: 'An error occurred on the server.' });
  }
};

exports.getMyLikeStatus = async (req, res) => {
  try {
    const me = await authService.getMe(req.headers.authorization);
    if (!me?.userId) return res.status(403).json({ success: false, message: "Forbidden" });

    const blogId = +req.params.blogId;
    if (!Number.isInteger(blogId) || blogId <= 0)
      return res.status(400).json({ success: false, message: "Invalid blogId" });

    const [liked, count] = await Promise.all([
      likeService.hasUserLiked(Number(me.userId), Number(blogId)),
      likeService.countForBlog(Number(blogId)),
    ]);


    return res.status(200).json({ success: true, data: { liked, count } });
  } catch (err) {
    console.error(err);
    return res.status(500).json({ success: false, message: "Server error" });
  }
};


exports.toggleLike = async (req, res) => {
  try {
    const me = await authService.getMe(req.headers.authorization);
    if (!me?.userId) return res.status(403).json({ success: false, message: "Forbidden" });

    const blogId = +req.params.blogId;
    if (!Number.isInteger(blogId) || blogId <= 0)
      return res.status(400).json({ success: false, message: "Invalid blogId" });

    const blog = await blogService.getBlogById(blogId);
    if (!blog) return res.status(404).json({ success: false, message: "Blog not found" });

    const allowed = await followerService.isUserFollowedByMe(req.headers.authorization, blog.userId);
    if (!allowed) {
      return res.status(403).json({
        success: false,
        message: "Forbidden: You must follow the author before liking."
      });
    }

    const result = await likeService.toggleLike(Number(me.userId), Number(blogId));
    return res.status(200).json({ success: true, data: result });
  } catch (err) {
    console.error(err);
    return res.status(500).json({ success: false, message: "Server error" });
  }
};

exports.getBlogLikesCount = async (req, res) => {
  try {
    const blogId = +req.params.blogId;
    if (!Number.isInteger(blogId) || blogId <= 0) {
      return res.status(400).json({ success: false, message: "Invalid blogId" });
    }
    const count = await likeService.countForBlog(Number(blogId));
    return res.status(200).json({ success: true, data: { count } });
  } catch (err) {
    console.error(err);
    return res.status(500).json({ success: false, message: "Server error" });
  }
};