const blogRepository = require('../repository/blogRepository');

exports.createBlog = async (blogData) => {
    
  const newBlog = await blogRepository.create(blogData);
  
  return newBlog;
};

exports.getAllBlogs = async () => {
  return await blogRepository.getAll();
};

exports.getBlogs = async (skip, limit, followed) => {
  return await blogRepository.getBlogs(skip, limit, followed);
};

exports.getBlogById = async (blogId) => {
  return await blogRepository.getById(blogId);
};