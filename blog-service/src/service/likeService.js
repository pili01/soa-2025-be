const likeRepository = require('../repository/likeRepository');

exports.createLike = async (likeData) => {
  return await likeRepository.create(likeData);
};

exports.getAllBlogLikes = async (blogId) => {
  return await likeRepository.getAllBlogLikes(blogId);
};

// Obrati pažnju na redosled: delete(userId, blogId)
exports.delete = async (blogId, userId) => {
  return await likeRepository.delete(userId, blogId);
};

exports.countForBlog = (blogId) => likeRepository.countByBlog(blogId);

exports.hasUserLiked = async (userId, blogId) => {
  const row = await likeRepository.findByUserAndBlog(userId, blogId);
  return !!row;
};

exports.toggleLike = async (userId, blogId) => {
  const existing = await likeRepository.findByUserAndBlog(userId, blogId);
  if (existing) {
    await likeRepository.delete(userId, blogId);
    const count = await likeRepository.countByBlog(blogId);
    return { liked: false, count };
  }
  await likeRepository.create({ userId, blogId });
  const count = await likeRepository.countByBlog(blogId);
  return { liked: true, count };
};
