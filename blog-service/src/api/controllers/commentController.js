const commentService = require('../../service/commentService');
const authService = require('../../service/authService');

exports.createComment = async (req, res, next) => {
    try {
        const data = await authService.getMe(req.headers.authorization);
        if (!data || !data.role || !data.userId) {
            return res.status(403).json({
                success: false,
                message: "Forbidden"
            });
        }

        if (data.role != "Tourist" && data.role != "Guide") {
            return res.status(403).json({
                success: false,
                message: "Forbidden"
            });
        }
        const commentData = req.body;
        commentData.userId = data.userId;
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