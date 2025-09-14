const blogService = require('../../service/blogService');
const authService = require('../../service/authService');
const followerService = require('../../service/followerService');

exports.createBlog = async (req, res, next) => {
  try {
    const blogData = req.body;
    console.log('Primljeni podaci:', blogData);
    const data = await authService.getMe(req.headers.authorization);
    if (!data || !data.role || !data.userId) {
      return res.status(403).json({
        success: false,
        message: "Forbidden"
      });
    }

    if (data.role != "Tourist" && data.role != "Guide") {
      return res.status(403).json({
        success: false,
        message: "Forbidden"
      });
    }
    blogData.userId = data.userId;
    const newBlog = await blogService.createBlog(blogData);

    res.status(201).json({
      success: true,
      data: newBlog,
    });
  } catch (error) {
    console.error(error);

    res.status(500).json({
      success: false,
      message: 'Došlo je do greške na serveru.',
    });
  }
};

exports.getAllBlogs = async (req, res) => {
  try {
    const blogs = await blogService.getAllBlogs();
    res.status(200).json({ success: true, data: blogs });
  } catch (error) {
    res.status(500).json({
      success: false,
      message: "There is an error that occured. ",
    });
  }
};

exports.getBlogs = async (req, res) => {
  try {
    const page = parseInt(req.query.page) || 1;
    const limit = parseInt(req.query.limit) || 10;

    const skip = (page - 1) * limit;
    var followed = await followerService.getFollowedUsers(req.headers.authorization);
    followed.push({id: (await authService.getMe(req.headers.authorization)).userId});
    console.log('Followed users:', followed.map(user => user.id));

    const blogs = await blogService.getBlogs(skip, limit, followed.map(user => user.id));
    res.status(200).json({ success: true, data: blogs });
  } catch (error) {
    res.status(500).json({
      success: false,
      message: "There is an error that occured. ",
    });
  }
}

exports.getBlogById = async (req, res) => {
  try {
    const raw = req.params.id; 
    const id = Number.parseInt(raw, 10);

    if (!Number.isInteger(id) || id <= 0) {
      return res.status(400).json({ success: false, message: 'Invalid blog id.' });
    }

    const blog = await blogService.getBlogById(id);

    if (!blog) {
      return res.status(404).json({
        success: false,
        message: 'Blog not found.'
      });
    }

    return res.status(200).json({
      success: true,
      data: blog
    });
  } catch (error) {
    console.error(error);
    return res.status(500).json({
      success: false,
      message: 'Došlo je do greške na serveru.'
    });
  }
};