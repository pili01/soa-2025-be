const commentService = require('../../service/commentService');

exports.createComment = async (req, res, next) => {
    try {
        const commentData = req.body;
        const savedComment = await commentService.createComment(commentData);
        res.status(201).json({
            success: true,
            data: savedComment,
        })
    } catch (error) {
        next(error);
    }
}

exports.getCommentsByBlogId = async (req, res, next) => {
    try {
        const blogId = req.params.blogId;
        const comments = await commentService.getCommentsByBlogId(+blogId);
        res.status(200).json({
            success: true,
            data: comments,
        });
    } catch (error) {
        next(error);
    }
}