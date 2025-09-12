const commentRepository = require('../repository/commentRepository');

exports.createComment = async (commentData) => {
    return await commentRepository.createComment(commentData);
}

exports.getCommentsByBlogId = async (blogId) => {
    return await commentRepository.getCommentsByBlogId(blogId);
}

exports.updateComment = async (commentId, newContent) => {
    return await commentRepository.updateComment(commentId, newContent);
}

exports.getCommentById = async (commentId) => {
    return await commentRepository.getCommentById(commentId);
}

exports.deleteComment = async (commentId) => {
    return await commentRepository.deleteComment(commentId);
}