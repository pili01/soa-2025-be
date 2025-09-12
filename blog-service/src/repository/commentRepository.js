const prisma = require('../config/prisma');

exports.createComment = async (commentData) => {
    return await prisma.BlogComment.create({
        data: commentData
    });
}

exports.getCommentsByBlogId = async (blogId) => {
    return await prisma.BlogComment.findMany({
        where: {
            blogId: blogId
        },
        orderBy: {
            createdAt: 'desc'
        }
    });
}

exports.updateComment = async (commentId, newContent) => {
    return await prisma.BlogComment.update({
        where: { id: Number(commentId) },
        data: { content: newContent.content }
    });
}

exports.getCommentById = async (commentId) => {
    return await prisma.BlogComment.findUnique({
        where: { id: Number(commentId) }
    });
}

exports.deleteComment = async (commentId) => {
    return await prisma.BlogComment.delete({
        where: { id: Number(commentId) }
    });
}