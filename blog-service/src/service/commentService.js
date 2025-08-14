const commentRepository = require('../repository/commentRepository');

exports.createComment = async (commentData) => {
    return await commentRepository.createComment(commentData);
}

exports.getCommentsByBlogId = async (blogId) => {
    return await commentRepository.getCommentsByBlogId(blogId);
}