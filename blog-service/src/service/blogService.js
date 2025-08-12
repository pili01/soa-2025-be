const blogRepository = require('../repository/blogRepository');

exports.createBlog = async (blogData) => {

  const newBlog = await blogRepository.create(blogData);
  
  return newBlog;
};