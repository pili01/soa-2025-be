const commentService = require('../../service/commentService');
const authService = require('../../service/authService');
const followerService = require('../../service/followerService');
const blogService = require('../../service/blogService');
const { blog } = require('../../config/prisma');

exports.createComment = async (req, res, next) => {
    try {
        const data = await authService.getMe(req.headers.authorization);
        if (!data || !data.role || !data.userId) {
            return res.status(403).json({
                success: false,
                message: "Forbidden"
            });
        }
        const blogId = req.body.blogId;
        if (!blogId) {
            return res.status(400).json({
                success: false,
                message: "Blog ID is required"
            });
        }
        const blog = await blogService.getBlogById(blogId);
        if (!blog) {
            return res.status(404).json({
                success: false,
                message: "Blog not found"
            });
        }
        console.log("Checking if user is followed...");
        if (!(await followerService.isUserFollowedByMe(req.headers.authorization, blog.userId)) && blog.userId !== data.userId) {
            return res.status(403).json({
                success: false,
                message: "Forbidden: You are not allowed to comment on this blog, you must follow the author."
            });
        }
        const commentData = req.body;
        commentData.userId = data.userId;
        commentData.authorUsername = data.username;
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

exports.updateComment = async (req, res, next) => {
    try {
        const data = await authService.getMe(req.headers.authorization);
        if (!data || !data.role || !data.userId) {
            return res.status(403).json({
                success: false,
                message: "Forbidden"
            });
        }
        const commentId = req.params.commentId;
        if (!commentId) {
            return res.status(400).json({
                success: false,
                message: "Comment ID is required"
            });
        }
        const comment = await commentService.getCommentById(commentId);
        if (!comment) {
            return res.status(404).json({
                success: false,
                message: "Comment not found"
            });
        }
        if (comment.userId !== data.userId) {
            return res.status(403).json({
                success: false,
                message: "Forbidden: You can only update your own comments."
            });
        }
        const updatedData = req.body;
        const updatedComment = await commentService.updateComment(commentId, updatedData);
        res.status(200).json({
            success: true,
            data: updatedComment,
        });
    } catch (error) {
        next(error);
    }
}

exports.deleteComment = async (req, res, next) => {
    try {
        const data = await authService.getMe(req.headers.authorization);
        if (!data || !data.role || !data.userId) {
            return res.status(403).json({
                success: false,
                message: "Forbidden"
            });
        }
        const commentId = req.params.commentId;
        if (!commentId) {
            return res.status(400).json({
                success: false,
                message: "Comment ID is required"
            });
        }
        const comment = await commentService.getCommentById(commentId);
        if (!comment) {
            return res.status(404).json({
                success: false,
                message: "Comment not found"
            });
        }
        if (comment.userId !== data.userId) {
            return res.status(403).json({
                success: false,
                message: "Forbidden: You can only delete your own comments."
            });
        }
        await commentService.deleteComment(commentId);
        res.status(204).json({
            success: true,
            message: "Comment deleted successfully"
        });
    } catch (error) {
        next(error);
    }
}
