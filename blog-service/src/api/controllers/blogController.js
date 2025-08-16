const blogService = require('../../service/blogService');

exports.createBlog = async (req, res, next) => {
  try {
    const blogData = req.body;
    console.log('Primljeni podaci:', blogData);

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