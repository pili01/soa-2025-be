const blogService = require('../../service/blogService');

exports.createBlog = async (req, res, next) => {
  try {
    const blogData = req.body;

    const newBlog = await blogService.createBlog(blogData);

    res.status(201).json({
      success: true,
      data: newBlog,
    });
  } catch (error) {
    console.error('Greška u createBlog kontroleru:', error);
    res.status(500).json({
      success: false,
      message: 'Došlo je do greške na serveru.',
    });
  }
};