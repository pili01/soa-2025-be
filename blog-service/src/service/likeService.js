const { blog } = require('../config/prisma');
const likeRepository = require('../repository/likeRepository');

exports.createLike = async(likeData) => {
    return await likeRepository.create(likeData);
} 

exports.getAllBlogLikes = async(blogId) => {
    return await likeRepository.getAllBlogLikes(blogId);
}

exports.delete = async(blogId, userId) => {
    return await likeRepository.delete(userId, blogId);
}