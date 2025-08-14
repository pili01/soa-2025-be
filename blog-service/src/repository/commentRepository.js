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
        }
    });
}